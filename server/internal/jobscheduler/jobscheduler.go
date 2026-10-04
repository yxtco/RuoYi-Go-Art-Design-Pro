package jobscheduler

import (
	"go-fin-server/internal/model"
	"go-fin-server/internal/service"
	"go-fin-server/pkg"
	"go-fin-server/pkg/taskregistry"
	"go-fin-server/pkg/types"
	"log"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

// TaskScheduler 定时任务调度器
type TaskScheduler struct {
	cron              *cron.Cron
	jobsMap           map[string]cron.EntryID
	jobsUpdateTimeMap map[string]types.LocalTime
	isReload          bool
}

// NewTaskScheduler 创建新的任务调度器
func NewTaskScheduler() *TaskScheduler {
	logger := cron.DefaultLogger
	return &TaskScheduler{
		cron:              cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(logger))),
		jobsMap:           make(map[string]cron.EntryID),
		jobsUpdateTimeMap: make(map[string]types.LocalTime),
		isReload:          true,
	}
}

// Start 启动任务调度器
func (ts *TaskScheduler) Start() {
	ts.cron.Start()
	jobs, err := service.Services{}.SysJobService.SelectJobAllList(map[string]interface{}{})
	if err != nil {
		pkg.Logger.Error(err.Error())
		return
	}
	for _, job := range jobs {
		ts.jobsUpdateTimeMap[job.JobName] = job.UpdatedAt
	}
	ts.cron.AddFunc("", func() {
		jobs, err := service.Services{}.SysJobService.SelectJobAllList(map[string]interface{}{})
		if err != nil {
			pkg.Logger.Error(err.Error())
		}
		for _, job := range jobs {
			ts.jobsUpdateTimeMap[job.JobName] = job.UpdatedAt
		}
	})
	ts.loadAndScheduleJobs(jobs)
	ts.isReload = false
}

// Stop 停止任务调度器
func (ts *TaskScheduler) Stop() {
	ts.cron.Stop()
}

func (ts *TaskScheduler) loadAndScheduleJobs(jobs []model.SysJob) {
	for _, job := range jobs {
		switch job.MisfirePolicy {
		case "1": // 立即执行
			ts.scheduleJob(job, false)
		case "2": // 执行一次
			ts.scheduleJob(job, true)
		default: // 3 放弃执行 上一次没有执行完成则放弃执行
			// 查询上一次执行完成时间 如果当前时间大于完成时间则执行
			jobLogList, _, err := model.SysJobLog{}.SelectJobLogList(map[string]interface{}{
				"job_name = ?": job.JobName,
			}, []string{"finish_time desc"}, 1, 1)
			if err != nil {
				pkg.Logger.Error(err.Error())
			}
			if len(jobLogList) > 0 {
				if time.Now().Second() <= jobLogList[0].FinishTime.Time.Second() {
					continue
				}
			}
			ts.scheduleJob(job, false)
		}
	}
}

func (ts *TaskScheduler) scheduleJob(job model.SysJob, isExecuteOnce bool) {
	if job.Status != "0" { // Skip if job is not active
		if entryID, ok := ts.jobsMap[job.JobName]; ok {
			ts.cron.Remove(entryID)
			delete(ts.jobsMap, job.JobName)
		}
		return
	}

	if !taskregistry.Exists(job.InvokeTarget) {
		log.Printf("Task function %s not found", job.InvokeTarget)
		return
	}
	if !ts.isReload {
		// 查询最新的调度日志 如果当前job未做调整则不执行
		jobLogList, err := model.SysJobLog{}.SelectJobLogAllList(map[string]interface{}{
			"job_name = ?": job.JobName,
		})
		if err != nil {
			pkg.Logger.Error(err.Error())
			return
		}
		if len(jobLogList) > 0 {
			if jobLogList[0].JobUpdatedTime == job.UpdatedAt {
				return
			}
		}
	}
	// 先移除
	if entryID, ok := ts.jobsMap[job.JobName]; ok {
		ts.cron.Remove(entryID)
		delete(ts.jobsMap, job.JobName)
	}
	// 再新增
	id, err := ts.cron.AddFunc(job.CronExpression, func() {
		startTime := time.Now()
		// 记录调度日志
		jobLog := model.SysJobLog{
			JobName:          job.JobName,
			JobGroup:         job.JobGroup,
			InvokeTarget:     job.InvokeTarget,
			InvokeTargetArgs: job.InvokeTargetArgs,
			JobUpdatedTime:   job.UpdatedAt,
		}
		err := jobLog.Create(&jobLog)
		if err != nil {
			pkg.Logger.Error(err.Error())
		}
		//// 是否并发执行
		//if job.Concurrent == "1"{
		//	go func(isExecuteOnce bool) {
		//
		//	}(isExecuteOnce)
		//}
		err = taskregistry.Execute(job.InvokeTarget, job.InvokeTargetArgs)
		if err != nil {
			duration := time.Since(startTime).Milliseconds()
			err := jobLog.UpdateMap(map[string]interface{}{
				"exception_info": err.Error(),
				"status":         "1",
				"job_message":    job.JobName + " 总共耗时：" + strconv.FormatInt(duration, 10) + "毫秒",
			}, "id = ?", jobLog.Id)
			if err != nil {
				pkg.Logger.Error(err.Error())
			}
			log.Printf("Error executing job %s: %v", job.JobName, err)
		} else {
			duration := time.Since(startTime).Milliseconds()
			err := jobLog.UpdateMap(map[string]interface{}{
				"job_message": job.JobName + " 总共耗时：" + strconv.FormatInt(duration, 10) + "毫秒",
			}, "id = ?", jobLog.Id)
			if err != nil {
				pkg.Logger.Error(err.Error())
			}
		}
		// 记录调度日志完成状态
		if isExecuteOnce {
			if entryID, ok := ts.jobsMap[job.JobName]; ok {
				ts.cron.Remove(entryID)
				delete(ts.jobsMap, job.JobName)
			}
		}
	})
	if err != nil {
		log.Printf("Error scheduling job %s: %v", job.JobName, err)
		return
	}

	ts.jobsMap[job.JobName] = id
}

// ReloadJobs 重新加载所有任务
func (ts *TaskScheduler) ReloadJobs() {
	for _, entryID := range ts.jobsMap {
		ts.cron.Remove(entryID)
	}
	ts.jobsMap = make(map[string]cron.EntryID)
	ts.jobsUpdateTimeMap = make(map[string]types.LocalTime)
	//ts.loadAndScheduleJobs()
}

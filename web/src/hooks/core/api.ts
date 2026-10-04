import { SysDictData } from '@/types/api/system/dict'
import response from '@utils/http'

export async function getDicts(dictType: string): Promise<SysDictData[]> {
  const res = await response<Record<string, SysDictData[]>>({
    url: '/system/dict/data/types',
    method: 'post',
    data: [dictType]
  })
  return res[dictType] || []
}

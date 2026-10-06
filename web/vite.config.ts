import path from 'path'
import { fileURLToPath } from 'url'

import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import ElementPlus from 'unplugin-element-plus/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import Components from 'unplugin-vue-components/vite'
import { defineConfig, loadEnv } from 'vite'
import viteCompression from 'vite-plugin-compression'
import vueDevTools from 'vite-plugin-vue-devtools'
// import { visualizer } from 'rollup-plugin-visualizer'

export default ({ mode }: { mode: string }) => {
  const root = process.cwd()
  const env = loadEnv(mode, root)
  const {
    VITE_VERSION,
    VITE_PORT,
    VITE_BASE_URL,
    VITE_API_URL,
    VITE_API_PROXY_URL
  } = env

  console.log(`🚀 API_URL = ${VITE_API_URL}`)
  console.log(`🚀 VERSION = ${VITE_VERSION}`)

  return defineConfig({
    define: {
      __APP_VERSION__: JSON.stringify(VITE_VERSION)
    },
    base: VITE_BASE_URL,
    server: {
      port: Number(VITE_PORT),
      proxy: {
        '/dev-api': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true,
          // 支持 WebSocket 升级转发（聊天 /ws/chat 走代理，避免跨域与主机/端口不一致）
          ws: true,
          rewrite: (p) => p.replace(/^\/dev-api/, ''),
          // 转发真实客户端 IP 到后端（登录日志/操作日志依赖 X-Real-IP / X-Forwarded-For）
          configure: (proxy) => {
            proxy.on('proxyReq', (proxyReq, req) => {
              const clientIP =
                req.headers['x-real-ip'] ||
                req.headers['x-forwarded-for']?.split(',')[0]?.trim() ||
                req.socket.remoteAddress
              if (clientIP) {
                proxyReq.setHeader('X-Real-IP', clientIP)
                proxyReq.setHeader(
                  'X-Forwarded-For',
                  req.headers['x-forwarded-for']
                    ? `${req.headers['x-forwarded-for']}, ${clientIP}`
                    : clientIP
                )
              }
            })
          }
        },
        // swaggo swagger doc.json 代理
        '/swagger/doc.json': {
          target: VITE_API_PROXY_URL,
          changeOrigin: true
        }
      },
      host: true
    },
    // 路径别名
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
        '@views': resolvePath('src/views'),
        '@imgs': resolvePath('src/assets/images'),
        '@icons': resolvePath('src/assets/icons'),
        '@utils': resolvePath('src/utils'),
        '@plugins': resolvePath('src/plugins'),
        '@stores': resolvePath('src/store'),
        '@styles': resolvePath('src/assets/styles')
      }
    },
    build: {
      target: 'es2015',
      outDir: '../server/web-dist', // 前端编译产物直接输出到 server/web-dist，后端启动即可托管
      assetsDir: 'assets-web',
      chunkSizeWarningLimit: 2000,
      minify: 'terser',
      terserOptions: {
        compress: {
          // 生产环境去除 console
          drop_console: true,
          // 生产环境去除 debugger
          drop_debugger: true
        }
      },
      dynamicImportVarsOptions: {
        warnOnError: true,
        exclude: [],
        include: ['src/views/**/*.vue']
      },
      rollupOptions: {
        output: {
          // 代码拆分：把大体积第三方库拆成独立 chunk，减小单个 JS 体积、提升并行加载
          manualChunks(id) {
            if (!id.includes('node_modules')) return undefined
            // element-plus 走按需导入，不手动合并（避免生成大 chunk）
            if (id.includes('/element-plus') || id.includes('@element-plus')) return undefined
            // Vue 生态合并为一个 chunk
            if (/[\\/]node_modules[\\/](@vue|vue|vue-router|pinia|vue-i18n|@vueuse)[\\/]/.test(id)) return 'vue'
            // 大体积库独立 chunk
            if (id.includes('echarts')) return 'echarts'
            if (id.includes('xlsx')) return 'xlsx'
            if (id.includes('wangeditor')) return 'wangeditor'
            if (id.includes('xgplayer')) return 'xgplayer'
            if (id.includes('axios')) return 'axios'
            return undefined
          }
        }
      }
    },
    plugins: [
      vue(),
      tailwindcss(),
      // 自动按需导入 API
      AutoImport({
        imports: ['vue', 'vue-router', 'pinia', '@vueuse/core'],
        dts: 'src/types/import/auto-imports.d.ts',
        resolvers: [ElementPlusResolver()]
        // eslintrc: {
        //   enabled: true,
        //   filepath: './.auto-import.json',
        //   globalsPropValue: true
        // }
      }),
      // 自动按需导入组件
      Components({
        dts: 'src/types/import/components.d.ts',
        resolvers: [ElementPlusResolver()]
      }),
      // 按需定制主题配置
      ElementPlus({
        useSource: true
      }),
      // 压缩（gzip + brotli），提升首屏加载速度
      // 产物同时保留原始文件与 .gz / .br，由后端按 Accept-Encoding 返回压缩版本
      viteCompression({
        verbose: false,
        disable: false,
        algorithm: 'gzip',          // gzip 压缩
        ext: '.gz',
        threshold: 1024,            // 大于 1KB 的资源均压缩（JS/CSS 全覆盖）
        deleteOriginFile: false
      }),
      viteCompression({
        verbose: false,
        disable: false,
        algorithm: 'brotliCompress', // brotli 压缩（比 gzip 更优）
        ext: '.br',
        threshold: 1024,
        deleteOriginFile: false
      }),
      vueDevTools()
      // 打包分析
      // visualizer({
      //   open: true,
      //   gzipSize: true,
      //   brotliSize: true,
      //   filename: 'dist/stats.html' // 分析图生成的文件名及路径
      // }),
    ],
    // 依赖预构建：避免运行时重复请求与转换，提升首次加载速度
    optimizeDeps: {
      include: [
        'echarts/core',
        'echarts/charts',
        'echarts/components',
        'echarts/renderers',
        'xlsx',
        'xgplayer',
        'crypto-js',
        'file-saver',
        'vue-cropper',
        'element-plus/es',
        'element-plus/es/components/*/style/css',
        'element-plus/es/components/*/style/index'
      ]
    },
    css: {
      preprocessorOptions: {
        // sass variable and mixin
        scss: {
          additionalData: `
            @use "@styles/core/el-light.scss" as *; 
            @use "@styles/core/mixin.scss" as *;
          `
        }
      },
      postcss: {
        plugins: [
          {
            postcssPlugin: 'internal:charset-removal',
            AtRule: {
              charset: (atRule) => {
                if (atRule.name === 'charset') {
                  atRule.remove()
                }
              }
            }
          }
        ]
      }
    }
  })
}

function resolvePath(paths: string) {
  return path.resolve(__dirname, paths)
}

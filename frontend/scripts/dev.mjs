import { spawn } from 'node:child_process'
import net from 'node:net'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const frontendDir = path.resolve(__dirname, '..')
const backendDir = path.resolve(frontendDir, '..', 'backend')

const BACKEND_HOST = '127.0.0.1'
const BACKEND_PORT = 8081
const BACKEND_READY_TIMEOUT_MS = 20000

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

const checkPort = (host, port, timeoutMs = 800) =>
  new Promise((resolve) => {
    const socket = new net.Socket()
    let done = false
    const finish = (result) => {
      if (done) return
      done = true
      socket.destroy()
      resolve(result)
    }
    socket.setTimeout(timeoutMs)
    socket.once('connect', () => finish(true))
    socket.once('timeout', () => finish(false))
    socket.once('error', () => finish(false))
    socket.connect(port, host)
  })

const waitBackendReady = async (timeoutMs) => {
  const start = Date.now()
  while (Date.now() - start < timeoutMs) {
    if (await checkPort(BACKEND_HOST, BACKEND_PORT)) return true
    await sleep(500)
  }
  return false
}

const spawnProc = (name, command, args, options) => {
  const child = spawn(command, args, {
    stdio: 'inherit',
    shell: true,
    ...options
  })
  child.on('exit', (code, signal) => {
    if (signal) {
      console.log(`[dev] ${name} exited by signal: ${signal}`)
    } else {
      console.log(`[dev] ${name} exited with code: ${code}`)
    }
  })
  return child
}

let backendProc = null
let viteProc = null
let shuttingDown = false

const shutdown = () => {
  if (shuttingDown) return
  shuttingDown = true

  if (viteProc && !viteProc.killed) viteProc.kill()
  if (backendProc && !backendProc.killed) backendProc.kill()

  setTimeout(() => process.exit(0), 200)
}

process.on('SIGINT', shutdown)
process.on('SIGTERM', shutdown)

const boot = async () => {
  const backendAlive = await checkPort(BACKEND_HOST, BACKEND_PORT)
  if (backendAlive) {
    console.log(`[dev] backend already running at http://${BACKEND_HOST}:${BACKEND_PORT}`)
  } else {
    console.log('[dev] starting backend...')
    backendProc = spawnProc('backend', 'go', ['run', '.'], {
      cwd: backendDir,
      env: {
        ...process.env,
        GOCACHE: path.join(backendDir, '.gocache')
      }
    })

    const ready = await waitBackendReady(BACKEND_READY_TIMEOUT_MS)
    if (!ready) {
      console.log('[dev] backend not ready within timeout, continuing with vite')
    } else {
      console.log(`[dev] backend ready at http://${BACKEND_HOST}:${BACKEND_PORT}`)
    }
  }

  viteProc = spawnProc('vite', 'vite', [], { cwd: frontendDir })
  viteProc.on('exit', () => shutdown())
}

boot().catch((err) => {
  console.error('[dev] failed to start dev environment:', err)
  shutdown()
})

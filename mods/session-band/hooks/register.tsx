import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import type { Summary } from '../types'

/**
 * session-band
 *
 * - At session start, a band above the prompt: where you left off on this
 *   project, read from the vault configured on THIS machine
 *   (~/.claude/vault.json), plus git status and open PRs. Hidden after the
 *   first prompt or with its button.
 * - /vault opens a pane with the last session note's closing state and next
 *   steps; /health re-runs the harness health check.
 * - A toast (and a native notification) when the session has been active for
 *   a while and nothing was written to the vault's sessions folder.
 */

const PANE = 'vault'
const summary = atom({ plugin: 'session-band', key: 'summary' } as const, null)
const isHidden = atom({ plugin: 'session-band', key: 'isHidden' } as const, false)

const REMIND_AFTER_MS = 3 * 60 * 60 * 1000 // 3h of session without a vault save
const REMIND_EVERY_MS = 60 * 60 * 1000 // then once an hour
const CHECK_EVERY_MS = 10 * 60 * 1000

type VaultConfig = {
  name?: string
  path?: string
  projects_dir?: string
  sessions_subdir?: string
  projects?: Record<string, string>
}

const basename = (p: string) => p.replace(/[\\/]+$/, '').split(/[\\/]/).pop() ?? p

const section = (md: string, heading: RegExp, max = 3): string[] => {
  const lines = md.split('\n')
  const start = lines.findIndex(l => heading.test(l))
  if (start < 0) return []
  const out: string[] = []
  for (const line of lines.slice(start + 1)) {
    if (/^#{1,6}\s/.test(line)) break
    const t = line.trim()
    if (t) out.push(t.replace(/^[-*]\s*(\[.\]\s*)?/, ''))
    if (out.length >= max) break
  }
  return out
}

let home = ''
let sessionsDir = ''

async function gather($: EngineInterface, cwd: string): Promise<Summary> {
    const project = basename(cwd)
    const s: Summary = {
      project, vaultOk: false, vaultNote: '', lastSessionDate: null, lastSessionTitle: null,
      daysSince: null, nextSteps: [], closingState: [], branch: '', dirty: 0, unpushed: 0, prs: [],
    }

    // ---- vault (this machine's config only)
    try {
      const cfg = JSON.parse(await $.fs.read(`${home}/.claude/vault.json`)) as VaultConfig
      if (!cfg.path || !(await $.fs.exists(cfg.path))) {
        s.vaultNote = cfg.path ? `vault no accesible: ${cfg.path}` : 'vault.json sin path'
      } else {
        const folder = cfg.projects?.[project] ?? `${cfg.path}/${cfg.projects_dir ?? '02-projects'}/${project}`
        if (!(await $.fs.exists(folder))) {
          s.vaultNote = `sin carpeta en el vault (${basename(cfg.path)})`
        } else {
          s.vaultOk = true
          sessionsDir = `${folder}/${cfg.sessions_subdir ?? 'sesiones'}`
          if (await $.fs.exists(sessionsDir)) {
            const notes = (await $.fs.list(sessionsDir))
              .filter(f => f.kind === 'file' && /^\d{4}-\d{2}-\d{2}-.*\.md$/.test(f.name))
              .sort((a, b) => b.name.localeCompare(a.name))
            const latest = notes[0]
            if (latest) {
              s.lastSessionDate = latest.name.slice(0, 10)
              const md = await $.fs.read(`${sessionsDir}/${latest.name}`)
              const h1 = md.split('\n').find(l => l.startsWith('# '))
              s.lastSessionTitle = h1 ? h1.replace(/^#\s*/, '').replace(/^\d{4}-\d{2}-\d{2}\s*[—-]\s*/, '') : latest.name
              s.nextSteps = section(md, /^##\s*pr[oó]ximo paso/i)
              s.closingState = section(md, /^##\s*estado al cerrar/i)
              const days = (Date.now() - new Date(s.lastSessionDate).getTime()) / 86400000
              s.daysSince = Math.max(0, Math.floor(days))
            } else {
              s.vaultNote = 'sin notas de sesion'
            }
          } else {
            s.vaultNote = 'sin carpeta sesiones/'
          }
        }
      }
    } catch {
      s.vaultNote = 'sin ~/.claude/vault.json'
    }

    // ---- git
    try {
      const st = await $.process.run(['git', 'status', '-sb', '--porcelain=v1'], { cwd, timeoutMs: 3000 })
      const lines = st.stdout.split('\n').filter(Boolean)
      if (st.exitCode === 0 && lines[0]?.startsWith('## ')) {
        const head = lines[0].slice(3)
        s.branch = head.split('...')[0].split(' ')[0]
        s.dirty = lines.length - 1
        const m = head.match(/\[ahead (\d+)/)
        s.unpushed = m ? Number(m[1]) : 0
      }
    } catch { /* not a repo */ }

    // ---- PRs (gh may be missing or offline: ignore)
    try {
      const pr = await $.process.run(['gh', 'pr', 'list', '--state', 'open', '--limit', '4', '--json', 'number,title'],
        { cwd, timeoutMs: 6000 })
      if (pr.exitCode === 0) {
        const list = JSON.parse(pr.stdout) as { number: number; title: string }[]
        s.prs = list.map(p => `#${p.number} ${p.title.slice(0, 40)}`)
      }
    } catch { /* no gh */ }

    return s
}

export const register: Register = (on) => {
  let startedAt = 0
  let lastRemindAt = 0
  let savedThisSession = false
  let turns = 0

  on('session.start', async ($, e, next) => {
    home = (await $.env.get('HOME')) ?? (await $.env.get('USERPROFILE')) ?? ''
    startedAt = await $.clock.now()
    await $.command.register({ name: 'vault', description: 'Pane: estado y proximo paso de la ultima sesion en el vault' })
    await $.command.register({ name: 'health', description: 'Re-run the harness health check (Obsidian, vault, hooks, agents)' })
    await update($, isHidden, () => false)

    void (async () => {
      const s = await gather($, e.cwd)
      await update($, summary, () => s)
    })()

    // Reminder: a long session with nothing written to the vault's sessions dir.
    $.clock.every(CHECK_EVERY_MS, async () => {
      if (savedThisSession || turns < 3) return
      const now = await $.clock.now()
      if (now - startedAt < REMIND_AFTER_MS) return
      if (now - lastRemindAt < REMIND_EVERY_MS) return
      lastRemindAt = now
      const hours = Math.floor((now - startedAt) / 3600000)
      const s = await read($, summary)
      const msg = `${hours}h de sesion en ${s?.project ?? 'este proyecto'} sin guardar en el vault: "anota todo en el vault" cuando cierres.`
      $.ui.toast(msg, { timeoutMs: 8000 })
      void $.ui.notify(msg, { title: 'session-save' }).catch(() => undefined)
    })

    return next(e)
  })

  on('prompt.submit', async ($, e, next) => {
    turns += 1
    await update($, isHidden, () => true)
    return next(e)
  })

  // A write inside the vault's sessions folder counts as "saved".
  on('tool.call', ($, e, next) => {
    if ((e.tool === 'Write' || e.tool === 'Edit') && sessionsDir && e.file_path.startsWith(sessionsDir)) {
      savedThisSession = true
    }
    return next(e)
  })

  on('command.run', { command: 'health' }, async $ => {
    const cwd = await $.session.root()
    const ran = await $.process.run(['python3', `${home}/.claude/scripts/session-health.py`],
      { cwd, env: { CLAUDE_PROJECT_DIR: cwd }, timeoutMs: 8000 }).catch(() => null)
    const text = ran && ran.exitCode === 0 ? ran.stdout.trim() : `health check no disponible (${ran?.stderr.trim() || 'script no encontrado'})`
    $.ui.toast(text.split('\n')[0] ?? text, { timeoutMs: 6000 })
    return { text }
  })

  on('command.run', { command: 'vault' }, async $ => {
    const s = await read($, summary)
    if (!s?.vaultOk) return { text: `Vault: ${s?.vaultNote ?? 'sin datos todavia'}` }
    const cwd = await $.session.root()
    await update($, summary, () => null)
    const fresh = await gather($, cwd)
    await update($, summary, () => fresh)
    await $.ui.open({ id: PANE, title: `Vault: ${fresh.project}` })
    return { text: 'Pane del vault abierto.' }
  })

  on('ui.render', { component: 'AbovePrompt' }, async ($, e, next) => {
    const s = await read($, summary)
    if (e.props.hasSurvey || s === null || (await read($, isHidden))) return next(e)

    const { Box, Button, Text } = $.ui.resolve(e)
    const stale = s.daysSince !== null && s.daysSince >= 3
    const last = s.lastSessionDate
      ? `ultima sesion ${s.lastSessionDate}${s.daysSince !== null ? ` (hace ${s.daysSince} d)` : ''}: ${s.lastSessionTitle ?? ''}`
      : `vault: ${s.vaultNote}`
    const git = s.branch
      ? `git ${s.branch}${s.dirty ? ` · ${s.dirty} sin commitear` : ''}${s.unpushed ? ` · ↑${s.unpushed} sin push` : ''}`
      : ''
    const prs = s.prs.length ? `PRs: ${s.prs.join(' · ')}` : ''

    return (
      <Box flexDirection="column">
        <Box>
          <Text bold>{s.project}</Text>
          <Text dimColor> — </Text>
          <Text color={stale ? 'yellow' : undefined} dimColor={!stale}>{last}</Text>
          <Text> </Text>
          <Button key="hide" label="Ocultar" onPress={() => update($, isHidden, () => true)} />
        </Box>
        {s.nextSteps.length > 0 && (
          <Text dimColor>→ proximo: {s.nextSteps.join(' · ')}</Text>
        )}
        {(git || prs) && (
          <Text dimColor>{[git, prs].filter(Boolean).join('  ·  ')}</Text>
        )}
      </Box>
    )
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Box, Text } = $.ui.resolve(e)
    const s = await read($, summary)
    if (!s) return <Text dimColor>Leyendo el vault...</Text>
    return (
      <Box flexDirection="column">
        <Text bold>{s.project}</Text>
        <Text dimColor>{s.lastSessionDate ? `Ultima sesion: ${s.lastSessionDate} — ${s.lastSessionTitle ?? ''}` : s.vaultNote}</Text>
        {s.closingState.length > 0 && <Text> </Text>}
        {s.closingState.length > 0 && <Text bold>Estado al cerrar</Text>}
        {s.closingState.map((l, i) => <Text key={`c${i}`}>  - {l}</Text>)}
        {s.nextSteps.length > 0 && <Text> </Text>}
        {s.nextSteps.length > 0 && <Text bold>Proximo paso</Text>}
        {s.nextSteps.map((l, i) => <Text key={`n${i}`}>  - {l}</Text>)}
        <Text> </Text>
        <Text dimColor>{s.branch ? `git ${s.branch} · ${s.dirty} sin commitear · ↑${s.unpushed}` : 'sin git'}</Text>
        {s.prs.map((p, i) => <Text key={`p${i}`} dimColor>  {p}</Text>)}
      </Box>
    )
  })
}

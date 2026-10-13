import type { Register } from 'claude-code'

/**
 * guardrails: the rules CLAUDE.md states as text, made impossible to skip.
 *
 * Bash:  git push --force / -f (unless --force-with-lease), deleting main or
 *        master, rm -rf on / or $HOME, git reset --hard on main/master.
 * Edit/Write/MultiEdit: .env files (not .env.example), anything under a
 *        secrets/ folder, credentials.json, *token*.json, id_rsa and friends.
 *
 * Each guard judges before calling next and has a catch handler, so a failing
 * guard denies instead of letting the call through.
 */

const PROTECTED_FILE = /(^|[\\/])(\.env(\.[^.]+)?|credentials\.json|[^\\/]*token[^\\/]*\.json|id_(rsa|ed25519|ecdsa)(\.pub)?)$|(^|[\\/])secrets[\\/]/i
const ALLOWED_FILE = /\.env\.(example|sample|template)$/i

const FORCE_PUSH = /\bgit\s+push\b(?![^|;&]*--force-with-lease)[^|;&]*(\s--force\b|\s-f\b|\s-[a-zA-Z]*f[a-zA-Z]*\b)/
const DELETE_MAIN = /\bgit\s+(branch\s+-[dD]|push\s+[^\s]+\s+--delete|push\s+[^\s]+\s+:)\s*(main|master)\b/
const RESET_MAIN = /\bgit\s+checkout\s+(main|master)\b[^|;&]*&&[^|;&]*git\s+reset\s+--hard|\bgit\s+reset\s+--hard\s+origin\/(main|master)\b/
const RM_HOME = /\brm\s+(-[a-zA-Z]*r[a-zA-Z]*f[a-zA-Z]*|-[a-zA-Z]*f[a-zA-Z]*r[a-zA-Z]*)\s+("?~"?|\$HOME|\/|\/\*|~\/\*|\$HOME\/\*)(\s|$)/

export const register: Register = (on) => {
  const fileGuard = (path: string) =>
    PROTECTED_FILE.test(path) && !ALLOWED_FILE.test(path)
      ? { deny: `guardrails: ${path} es un archivo sensible; editalo vos a mano.` }
      : null

  on('tool.call', { tool: 'Edit' }, ($, e, next) => fileGuard(e.file_path) ?? next(e))
    .catch(($, e, next) => (next.called ? next(e) : { deny: 'guardrails: el guard fallo.' }))

  on('tool.call', { tool: 'Write' }, ($, e, next) => fileGuard(e.file_path) ?? next(e))
    .catch(($, e, next) => (next.called ? next(e) : { deny: 'guardrails: el guard fallo.' }))

  on('tool.call', { tool: 'Bash' }, ($, e, next) => {
    const cmd = e.command
    if (FORCE_PUSH.test(cmd)) return { deny: 'guardrails: no force-push (usa --force-with-lease si de verdad hace falta).' }
    if (DELETE_MAIN.test(cmd)) return { deny: 'guardrails: no se borra main/master.' }
    if (RESET_MAIN.test(cmd)) return { deny: 'guardrails: no reset --hard sobre main/master.' }
    if (RM_HOME.test(cmd)) return { deny: 'guardrails: rm -rf sobre home o raiz, no.' }
    return next(e)
  }).catch(($, e, next) => (next.called ? next(e) : { deny: 'guardrails: el guard fallo.' }))
}

import type { EngineInterface, Register } from 'claude-code'

/**
 * guardrails: the rules CLAUDE.md states as text, made impossible to skip,
 * plus two evidence guards.
 *
 * Hard denies
 *   Bash:  git push --force / -f (unless --force-with-lease), deleting main or
 *          master, rm -rf on / or $HOME, git reset --hard on main/master.
 *   Edit/Write: .env files (not .env.example), anything under secrets/,
 *          credentials.json, *token*.json, SSH private keys.
 *
 * Evidence guards
 *   1. No tests, no commit: when code files were edited and no verification
 *      command (tests, linter, type-check, terraform validate...) ran
 *      successfully afterwards, `git commit` is denied. Only in repos that have
 *      a runner (pyproject/pytest.ini, package.json "test", go.mod, Cargo.toml,
 *      Makefile test target). Escape hatch: prefix the command with
 *      `SKIP_GUARD=1` when the change really needs no tests.
 *   2. Claims without evidence: when the final answer of a turn says things
 *      like "listo", "funciona", "tests en verde", "verified" and no
 *      verification command ran in that turn, a toast says so. It never blocks.
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

// What counts as verification: tests, linters, type checks, validators.
const VERIFY_CMD = /\b(pytest|python3?\s+-m\s+(pytest|unittest|py_compile)|npm\s+(run\s+)?(test|lint|typecheck)|pnpm\s+(run\s+)?(test|lint)|yarn\s+(test|lint)|npx\s+(vitest|jest|eslint|tsc|prettier\s+--check)|vitest|jest|eslint|tsc\b|ruff\b|flake8|mypy|black\s+--check|go\s+(test|vet)|cargo\s+(test|clippy|check)|make\s+(test|lint|check)|bash\s+-n\b|shellcheck|terraform\s+(validate|plan)|tflint|helm\s+lint|kubectl\s+apply\s+--dry-run|docker\s+build|claude\s+plugin\s+(validate|test)|phpunit|mvn\s+test|gradle\s+test|dotnet\s+test)/
const GIT_COMMIT = /\bgit\s+(-C\s+\S+\s+)?commit\b/
const SKIP = /\bSKIP_GUARD=1\b/
const CODE_FILE = /\.(py|ts|tsx|js|jsx|mjs|cjs|go|rs|java|kt|rb|php|cs|sh|bash|zsh|fish|tf|hcl|yaml|yml|toml|json|sql|c|h|cpp|hpp|swift|dart|vue|svelte)$/i
const CLAIM = /\b(listo|funciona(n)?|anda(n)?|pasa(n)?\s+(los\s+|todos\s+los\s+)?tests?|tests?\s+(en\s+)?verde|todo\s+(ok|verde|bien)|verificad[oa]s?|probad[oa]s?|comprobad[oa]s?|works|working|passing|all\s+green|verified|tests?\s+pass(ed)?)\b/i

// --- evidence state (module variables: reset on reload, which is fine for a session)
const editedCode = new Map<string, number>() // code file -> when it was edited (ms)
let lastVerifyOkAt = 0
let verifyRanThisTurn = false
let editedThisTurn = false

function fileGuard(path: string) {
  return PROTECTED_FILE.test(path) && !ALLOWED_FILE.test(path)
    ? { deny: `guardrails: ${path} es un archivo sensible; editalo vos a mano.` }
    : null
}

async function noteEdit($: EngineInterface, path: string): Promise<void> {
  if (CODE_FILE.test(path)) {
    editedCode.set(path, await $.clock.now())
    editedThisTurn = true
  }
}

export const register: Register = (on) => {

  on('tool.call', { tool: 'Edit' }, async ($, e, next) => {
    const denied = fileGuard(e.file_path)
    if (denied) return denied
    await noteEdit($, e.file_path)
    return next(e)
  }).catch(($, e, next) => (next.called ? next(e) : { deny: 'guardrails: el guard fallo.' }))

  on('tool.call', { tool: 'Write' }, async ($, e, next) => {
    const denied = fileGuard(e.file_path)
    if (denied) return denied
    await noteEdit($, e.file_path)
    return next(e)
  }).catch(($, e, next) => (next.called ? next(e) : { deny: 'guardrails: el guard fallo.' }))

  on('tool.call', { tool: 'Bash' }, async ($, e, next) => {
    const cmd = e.command
    if (FORCE_PUSH.test(cmd)) return { deny: 'guardrails: no force-push (usa --force-with-lease si de verdad hace falta).' }
    if (DELETE_MAIN.test(cmd)) return { deny: 'guardrails: no se borra main/master.' }
    if (RESET_MAIN.test(cmd)) return { deny: 'guardrails: no reset --hard sobre main/master.' }
    if (RM_HOME.test(cmd)) return { deny: 'guardrails: rm -rf sobre home o raiz, no.' }

    // Guard 1: no tests, no commit.
    if (GIT_COMMIT.test(cmd) && !SKIP.test(cmd)) {
      const root = commitRoot(cmd, await $.session.root())
      const pending = [...editedCode.entries()].filter(([f, at]) => f.startsWith(root) && at > lastVerifyOkAt)
      if (pending.length > 0 && (await hasRunner($, root))) {
        const sample = pending.slice(0, 3).map(([f]) => f.slice(root.length + 1)).join(', ')
        return {
          deny: `guardrails: ${pending.length} archivo(s) de codigo editados sin correr tests/lint despues (${sample}). ` +
            `Corre el runner del repo primero, o anteponele SKIP_GUARD=1 al commit si de verdad no aplica.`,
        }
      }
    }

    const ran = await next(e)
    if (VERIFY_CMD.test(cmd)) {
      verifyRanThisTurn = true
      if (ran.deny === undefined && ran.isError !== true) lastVerifyOkAt = await $.clock.now()
    }
    return ran
  }).catch(($, e, next) => (next.called ? next(e) : { deny: 'guardrails: el guard fallo.' }))

  on('prompt.submit', ($, e, next) => {
    verifyRanThisTurn = false
    editedThisTurn = false
    return next(e)
  })

  // Guard 2: a claim of success with no verification in the turn -> toast.
  on('turn.complete', async ($, e, next) => {
    const done = await next(e)
    if (e.agentId === undefined && e.reason === 'answer' && !verifyRanThisTurn && editedThisTurn && CLAIM.test(done.text ?? '')) {
      $.ui.toast('guardrails: afirmacion de exito sin ningun comando de verificacion en este turno.', { timeoutMs: 7000 })
    }
    return done
  })
}

/** The repo a `git commit` targets: `-C <dir>`, a leading `cd <dir> &&`, else the session root. */
function commitRoot(cmd: string, sessionRoot: string): string {
  const c = cmd.match(/\bgit\s+-C\s+("([^"]+)"|'([^']+)'|(\S+))\s+commit/)
  const cd = cmd.match(/^\s*cd\s+("([^"]+)"|'([^']+)'|(\S+))\s*&&/)
  const dir = c ? (c[2] ?? c[3] ?? c[4]) : cd ? (cd[2] ?? cd[3] ?? cd[4]) : sessionRoot
  return dir.replace(/\/+$/, '')
}

/** Whether the repo has something that counts as a test/lint runner. */
async function hasRunner($: EngineInterface, root: string): Promise<boolean> {
  const markers = ['pyproject.toml', 'pytest.ini', 'setup.cfg', 'tox.ini', 'go.mod', 'Cargo.toml', 'Makefile', 'package.json', 'ruff.toml', '.eslintrc.json', 'tsconfig.json']
  for (const m of markers) {
    if (!(await $.fs.exists(`${root}/${m}`))) continue
    if (m === 'package.json') {
      try {
        const pkg = JSON.parse(await $.fs.read(`${root}/package.json`)) as { scripts?: Record<string, string> }
        if (pkg.scripts && (pkg.scripts.test || pkg.scripts.lint || pkg.scripts.typecheck)) return true
      } catch { /* malformed: ignore */ }
      continue
    }
    if (m === 'Makefile') {
      try {
        if (/^(test|lint|check):/m.test(await $.fs.read(`${root}/Makefile`))) return true
      } catch { /* ignore */ }
      continue
    }
    return true
  }
  return false
}

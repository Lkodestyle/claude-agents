export type Summary = {
  project: string
  vaultOk: boolean
  vaultNote: string
  lastSessionDate: string | null
  lastSessionTitle: string | null
  daysSince: number | null
  nextSteps: string[]
  closingState: string[]
  branch: string
  dirty: number
  unpushed: number
  prs: string[]
}

declare module 'claude-code' {
  interface PluginState {
    'session-band': { summary: Summary | null; isHidden: boolean }
  }
}

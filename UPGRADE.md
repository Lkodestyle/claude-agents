# Upgrade Guide

How to update your Claude Agents installation without losing your existing memory or configuration.

## What Gets Preserved

```
~/.claude/
├── memory.json              # ✅ YOUR MEMORY - NEVER TOUCHED
├── router/                  # ✅ Router state per project - preserved
├── attention_history.jsonl  # ✅ Your history - preserved
├── pool/                    # ✅ Your pool state - preserved
├── settings.local.json      # ✅ Your local settings - preserved
├── agents/                  # 🔄 Our 15 files updated; anything else you keep there is preserved
├── scripts/                 # 🔄 Our scripts updated; other tools' scripts preserved
├── skills/                  # 🔄 Our 6 skills updated; other skills (plugins, your own) preserved
├── keywords.json            # 🔄 Updated
├── settings.json            # 🔄 MERGED: only our 3 hooks + permissions; model/theme/other hooks preserved
└── backups/claude-agents-*/ # 📦 Files of ours you had modified locally, moved here before updating
```

Agents, scripts and skills are installed **entry by entry** (one
symlink or copy per file/folder), never as a whole-folder replacement.

## Quick Upgrade

Same command for symlink and copy installs:

```bash
cd ~/claude-agents
git pull
./scripts/claude-agents-cli.sh sync     # detects the mode you installed with
./scripts/claude-agents-cli.sh status   # also checks hooks are not duplicated
```

`sync` re-links (or re-copies) new entries, prunes entries removed upstream,
and re-merges the hooks into `~/.claude/settings.json`.

### Upgrading from a version that symlinked whole folders

Older installs made `~/.claude/agents` (etc.) a symlink to the repo. Just run
`install` again: it replaces each folder symlink with a real folder and
per-entry links, and merges the hooks into your existing `settings.json`
instead of overwriting it.

```bash
./scripts/claude-agents-cli.sh install            # --symlink (default) or --copy
```

If your `~/.claude/settings.json` has hook `timeout` values like `5000`, they
came from an old version that assumed milliseconds; Claude Code reads them as
**seconds**. `install`/`sync` rewrites our hooks with correct values (10/5/5).

### Manual upgrade (full control)

```bash
# 1. Update repo
cd ~/claude-agents
git pull

# 2. Backup your memory (optional but recommended)
cp ~/.claude/memory.json ~/.claude/memory.backup.json

# 3. Update agents, scripts and skills (file by file, keeps your own)
cp -r ~/claude-agents/.claude/agents/*   ~/.claude/agents/
cp -r ~/claude-agents/.claude/scripts/*  ~/.claude/scripts/
cp -r ~/claude-agents/.claude/skills/*   ~/.claude/skills/

# 4. Update keywords
cp ~/claude-agents/.claude/keywords.json ~/.claude/

# 5. Hooks: do NOT copy .claude/settings.json over yours (it has no hooks and
#    would drop your model/theme). Merge them with the installer instead:
./scripts/claude-agents-cli.sh sync

# 6. Verify memory is intact
python3 ~/.claude/scripts/memory-manager.py stats
```

## Jarvis Moved to Its Own Repo

Jarvis se movió a su propio repo (https://github.com/Lkodestyle/jarvis); si tenías `bin/jarvis-*` o el MCP `jarvis-memory` registrado, ver el README del nuevo repo.

## Verify Upgrade

```bash
# Check installation status
./scripts/claude-agents-cli.sh status

# Test cognitive features
./scripts/claude-agents-cli.sh test

# Check your memory is intact
python3 ~/.claude/scripts/memory-manager.py stats

# Expected output:
# [✓] Claude home: /home/user/.claude
# [✓] agents: symlinked
# [✓] scripts: symlinked
# [✓] skills: symlinked
# [✓] context-router.py: global
# [✓] pool-loader.py: global
# [✓] pool-extractor.py: global
# [✓] Keywords config: present
```

If a hook line shows `global=1 project=1 -> would run 2 times`, a project
`.claude/settings.json` declares the same hook as your global settings: remove
it from the project file (hooks belong in the global one).

## What's New (v2.0)

After upgrading, you'll have access to:

### New Agents (+6)
- `kubernetes` - K8s, Helm, troubleshooting
- `docker` - Dockerfiles, multi-stage builds
- `monitoring` - Prometheus, Grafana, alerting
- `finops` - Cost optimization, rightsizing
- `security` - CVE scanning, secrets detection
- `mobile` - React Native, Expo, Flutter

### Slash Commands (removed)
The 8 slash commands were removed: never used in practice, and two shadowed built-ins.
`install`/`sync` delete the old links from `~/.claude/commands/`. Your own commands there are kept.

### Cognitive Features
- **Context Router** - Auto-activates agents based on keywords
- **Pool Coordinator** - Multi-instance coordination
- **Memory Manager** - Manage MCP memory size

## Memory Management

If you're getting token errors or memory is too large:

```bash
# Check memory size
python3 ~/.claude/scripts/memory-manager.py stats

# If > 20,000 tokens, consider cleaning
python3 ~/.claude/scripts/memory-manager.py export ~/memory-backup.json
python3 ~/.claude/scripts/memory-manager.py clear

# Or search and selectively clean
python3 ~/.claude/scripts/memory-manager.py search "old-project"
```

## Troubleshooting

### Hooks not working

Check settings.json exists and has correct format:

```bash
cat ~/.claude/settings.json
```

Should contain:
```json
{
  "hooks": {
    "UserPromptSubmit": [...],
    "SessionStart": [...],
    "Stop": [...]
  }
}
```

### Context router not activating agents

```bash
# Test manually
echo '{"prompt":"help with kubernetes deployment"}' | python3 ~/.claude/scripts/context-router.py

# Should output agent context with kubernetes activated
```

### Memory manager can't find memory

```bash
# Check memory location
ls -la ~/.claude/memory.json

# Or set custom path
export MCP_MEMORY_PATH=~/.claude/memory.json
```

### Pool not syncing between instances

```bash
# Ensure pool directory exists
mkdir -p ~/.claude/pool

# Set instance ID
export CLAUDE_INSTANCE=A  # or B, C, etc.

# Check pool state
python3 ~/.claude/scripts/pool-query.py --count
```

## Rollback

If something goes wrong:

```bash
# Restore memory from backup
cp ~/.claude/memory.backup.json ~/.claude/memory.json

# Or reinstall fresh (keeps memory)
./scripts/claude-agents-cli.sh uninstall
./scripts/claude-agents-cli.sh install
```

## Environment Variables

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
# Optional: Instance ID for pool coordination
export CLAUDE_INSTANCE=A

# Optional: Notion integration
export NOTION_TOKEN="ntn_your_token"

# Optional: Increase MCP token limit if needed
export MAX_MCP_OUTPUT_TOKENS=250000
```

## Questions?

- Check `./scripts/claude-agents-cli.sh help`
- Run `./scripts/claude-agents-cli.sh status` for diagnostics
- See README.md for full documentation

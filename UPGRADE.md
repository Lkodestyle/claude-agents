# Upgrade Guide

How to update your Claude Agents installation without losing your existing memory or configuration.

## What Gets Preserved

```
~/.claude/
├── memory.json              # ✅ YOUR MEMORY - NEVER TOUCHED
├── attn_state.json          # ✅ Your attention state - preserved
├── attention_history.jsonl  # ✅ Your history - preserved
├── pool/                    # ✅ Your pool state - preserved
├── settings.local.json      # ✅ Your local settings - preserved
├── agents/                  # 🔄 Updated
├── scripts/                 # 🔄 Updated
├── commands/                # 🔄 Updated
├── keywords.json            # 🔄 Updated
└── settings.json            # 🔄 Updated (hooks config)
```

## Quick Upgrade

### If you used symlinks (default installation)

```bash
# Just pull the latest changes
cd ~/claude-agents
git pull

# Copy new config files (if they don't exist)
[ ! -f ~/.claude/keywords.json ] && cp ~/claude-agents/.claude/keywords.json ~/.claude/
[ ! -f ~/.claude/settings.json ] && cp ~/claude-agents/.claude/settings.json ~/.claude/

# Create pool directory
mkdir -p ~/.claude/pool

# Verify
./scripts/claude-agents-cli.sh status
```

### If you used --copy installation

```bash
# Update repo
cd ~/claude-agents
git pull

# Sync files (preserves memory)
./scripts/claude-agents-cli.sh sync
```

### Manual upgrade (full control)

```bash
# 1. Update repo
cd ~/claude-agents
git pull

# 2. Backup your memory (optional but recommended)
cp ~/.claude/memory.json ~/.claude/memory.backup.json

# 3. Update agents (new and updated)
cp -r ~/claude-agents/.claude/agents/* ~/.claude/agents/

# 4. Add scripts (new)
cp -r ~/claude-agents/.claude/scripts ~/.claude/

# 5. Add commands (new)
cp -r ~/claude-agents/.claude/commands ~/.claude/

# 6. Update configs
cp ~/claude-agents/.claude/keywords.json ~/.claude/
cp ~/claude-agents/.claude/settings.json ~/.claude/

# 7. Create pool directory
mkdir -p ~/.claude/pool

# 8. Verify memory is intact
python3 ~/.claude/scripts/memory-manager.py stats
```

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
# [✓] Agents: symlinked -> /home/user/claude-agents/.claude/agents
# [✓] Scripts: symlinked
# [✓] Keywords config: present
# [✓] Memory: X entities, Y relations
```

## What's New (v2.0)

After upgrading, you'll have access to:

### New Agents (+6)
- `kubernetes` - K8s, Helm, troubleshooting
- `docker` - Dockerfiles, multi-stage builds
- `monitoring` - Prometheus, Grafana, alerting
- `finops` - Cost optimization, rightsizing
- `security` - CVE scanning, secrets detection
- `mobile` - React Native, Expo, Flutter

### Slash Commands (+8)
- `/commit` - Generate conventional commit
- `/pr` - Create pull request
- `/review` - Code review
- `/test` - Generate tests
- `/explain` - Explain code
- `/refactor` - Refactoring suggestions
- `/debug` - Debug errors
- `/doc` - Generate documentation

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

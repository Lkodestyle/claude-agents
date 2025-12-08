# Claude DevOps Agents 🤖

Colección de agentes especializados para Claude Code, enfocados en DevOps, Cloud Infrastructure y desarrollo.

## 🚀 Quick Start

### Instalación automática
```bash
git clone https://github.com/Lkodestyle/claude-agents.git ~/claude-agents
cd ~/claude-agents
./scripts/install.sh
```

### Instalación manual
```bash
git clone https://github.com/Lkodestyle/claude-agents.git ~/claude-agents
mkdir -p ~/.claude
ln -s ~/claude-agents/.claude/agents ~/.claude/agents
```

## 📦 Agentes Disponibles

| Agente | Especialidad |
|--------|--------------|
| 🏗️ `architecture` | Diseño de sistemas, C4, ADRs, trade-offs |
| 🔧 `terraform` | Terraform, Terraspace, Terragrunt, IaC |
| ☁️ `aws` | Amazon Web Services (EC2, ECS, RDS, etc.) |
| 🔵 `azure` | Microsoft Azure (App Services, AKS, etc.) |
| 🌐 `networking` | VPCs, subnets, CIDR, security groups |
| 🗄️ `databases` | PostgreSQL, DynamoDB, Redis, migrations |
| 💻 `programming` | Clean code, patterns, testing, code review |
| 🌍 `web` | React, Node.js, Docker, APIs, Nginx |
| 🔄 `cicd` | GitHub Actions, GitLab CI, pipelines |

## 📁 Estructura

```
claude-agents/
├── CLAUDE.md                    # Instrucciones principales
├── .claude/
│   └── agents/                  # Subagentes nativos de Claude Code
│       ├── architecture.md
│       ├── terraform.md
│       ├── aws.md
│       ├── azure.md
│       ├── networking.md
│       ├── databases.md
│       ├── programming.md
│       ├── web.md
│       └── cicd.md
├── agents/                      # Versiones originales (referencia)
├── templates/                   # Templates reutilizables
├── scripts/
│   └── install.sh              # Script de instalación
└── mcp-config.json             # Config de MCP servers
```

## 🔄 Uso

### Automático
Claude Code detectará automáticamente los agentes y los usará según el contexto.

### Explícito
Podés pedir usar un agente específico:
- "Usa el agente de terraform para crear un módulo de VPC"
- "Necesito al especialista de AWS para revisar esta arquitectura"

### Ver agentes
Ejecutá `/agents` en Claude Code para ver y gestionar los agentes disponibles.

## 🔄 Sincronización entre PCs

```bash
# En cualquier PC, actualizar agentes
cd ~/claude-agents
git pull
```

## 📝 Personalización

Podés crear `CLAUDE.local.md` en cualquier proyecto para agregar contexto específico:

```markdown
# Contexto Local (no se commitea)

- Cliente: MiEmpresa
- Ambiente: Production US-East-1
```

## 🤝 Contribuir

1. Fork el repo
2. Crea tu branch: `git checkout -b feature/nuevo-agente`
3. Commit: `git commit -m 'Agregar agente de Kubernetes'`
4. Push: `git push origin feature/nuevo-agente`
5. Abre un PR

## 📜 License

MIT

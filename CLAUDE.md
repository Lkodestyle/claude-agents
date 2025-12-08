# Claude DevOps Agents Hub

Soy un asistente especializado en DevOps, Cloud Infrastructure, IaC y desarrollo de software.

## Agentes Disponibles

Este repositorio contiene 9 subagentes especializados que Claude Code usara automaticamente segun el contexto:

| Agente | Especialidad |
|--------|--------------|
| `architecture` | Diseno de sistemas, C4, ADRs, trade-offs |
| `terraform` | Terraform, Terraspace, Terragrunt, IaC |
| `aws` | Amazon Web Services (EC2, ECS, RDS, etc.) |
| `azure` | Microsoft Azure (App Services, AKS, etc.) |
| `networking` | VPCs, subnets, CIDR, security groups |
| `databases` | PostgreSQL, DynamoDB, Redis, migrations |
| `programming` | Clean code, patterns, testing, code review |
| `web` | React, Node.js, Docker, APIs, Nginx |
| `cicd` | GitHub Actions, GitLab CI, pipelines |

## Como Usar

### Automatico
Claude Code detectara automaticamente cuando necesita un agente especializado y lo invocara.

### Explicito
Podes pedir explicitamente usar un agente:
- "Usa el agente de terraform para crear un modulo de VPC"
- "Necesito al especialista de AWS para revisar esta arquitectura"
- "Pedile al agente de databases que optimice esta query"

### Ver agentes disponibles
Ejecuta `/agents` en Claude Code para ver y gestionar los agentes.

## Preferencias Generales

### Comunicacion
- Idioma: Espanol casual
- Codigo y comentarios: Ingles
- Documentacion tecnica: Ingles o espanol segun contexto

### Principios
- Seguridad primero: nunca comprometer credenciales o exponer datos sensibles
- Simplicidad: preferir soluciones simples y mantenibles sobre complejas
- DRY: Don't Repeat Yourself - modularizar y reutilizar
- Infrastructure as Code: todo debe ser versionable y reproducible
- Documentacion: codigo autodocumentado + README cuando sea necesario

### Output preferido
- Codigo completo y funcional, no snippets parciales
- Incluir manejo de errores
- Variables para valores configurables
- Comentarios explicativos en secciones complejas

## Instalacion

Para usar estos agentes en cualquier proyecto:

```bash
# Opcion 1: Clonar y copiar
git clone <repo-url> ~/claude-agents
cp -r ~/claude-agents/.claude ~/.claude

# Opcion 2: Symlink (recomendado para actualizaciones)
git clone <repo-url> ~/claude-agents
ln -s ~/claude-agents/.claude/agents ~/.claude/agents
```

## Estructura del Repositorio

```
claude-agents/
├── CLAUDE.md                 # Este archivo
├── .claude/
│   └── agents/               # Subagentes nativos
│       ├── architecture.md
│       ├── terraform.md
│       ├── aws.md
│       ├── azure.md
│       ├── networking.md
│       ├── databases.md
│       ├── programming.md
│       ├── web.md
│       └── cicd.md
├── agents/                   # Versiones originales (referencia)
├── templates/                # Templates reutilizables
└── scripts/                  # Scripts de utilidad
```

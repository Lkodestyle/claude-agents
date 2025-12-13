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

### Automatico (Proactivo)
Los agentes tienen frases `USE PROACTIVELY` y `MUST BE USED` en sus descripciones, lo que indica a Claude Code cuando invocarlos automaticamente. Por ejemplo:
- Al escribir codigo Terraform -> se activa `terraform`
- Al trabajar con recursos AWS -> se activa `aws`
- Al disenar arquitectura de sistemas -> se activa `architecture`

### Explicito
Tambien podes pedir explicitamente usar un agente:
- "Usa el agente de terraform para crear un modulo de VPC"
- "Necesito al especialista de AWS para revisar esta arquitectura"
- "Pedile al agente de databases que optimice esta query"

### Ver agentes disponibles
Ejecuta `/agents` en Claude Code para ver y gestionar los agentes.

### Nota Importante
Si bien los agentes estan configurados para uso proactivo, Claude Code no siempre los invoca automaticamente en todos los casos. Para garantizar su uso, es mejor mencionarlos explicitamente cuando los necesites.

## MCP Servers Incluidos

Este repositorio incluye configuracion de MCP servers en `.mcp.json`:

| Server | Descripcion | Requiere Config |
|--------|-------------|-----------------|
| `memory` | Memoria persistente entre sesiones | No |
| `context7` | Documentacion actualizada de librerias (agrega "use context7" al prompt) | No |
| `supabase` | Interaccion con proyectos Supabase | OAuth automatico |
| `notion` | Acceso a workspaces de Notion | `NOTION_TOKEN` env var |

### Configuracion requerida

**Notion:** Necesitas crear un token de integracion:
1. Ve a https://www.notion.so/profile/integrations
2. Crea una nueva integracion interna
3. Configura la variable de entorno:
   ```bash
   # Agregar a ~/.bashrc o ~/.zshrc
   export NOTION_TOKEN="ntn_tu_token_aqui"
   ```

**Supabase:** Usa OAuth automatico, no requiere token manual.

**Context7:** Agrega "use context7" a tu prompt para obtener documentacion actualizada.

**Memory:** No requiere configuracion, funciona out-of-the-box.

Para verificar el estado de los MCP servers, ejecuta `/mcp` en Claude Code.

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
# Opcion 1: Clonar y copiar todo
git clone <repo-url> ~/claude-agents
cp -r ~/claude-agents/.claude ~/.claude
cp ~/claude-agents/.mcp.json ~/.mcp.json

# Opcion 2: Symlink (recomendado para actualizaciones)
git clone <repo-url> ~/claude-agents
ln -s ~/claude-agents/.claude/agents ~/.claude/agents

# Opcion 3: Copiar a un proyecto especifico
cp -r ~/claude-agents/.claude /mi-proyecto/.claude
cp ~/claude-agents/.mcp.json /mi-proyecto/.mcp.json
```

**Nota:** El archivo `.mcp.json` debe estar en la raiz del proyecto para que Claude Code lo detecte.

## Estructura del Repositorio

```
claude-agents/
├── CLAUDE.md                 # Este archivo
├── .mcp.json                 # Configuracion de MCP servers (memory)
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

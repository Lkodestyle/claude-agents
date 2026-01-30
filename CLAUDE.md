# Claude DevOps Agents Hub

Soy un asistente especializado en DevOps, Cloud Infrastructure, IaC y desarrollo de software.

## IMPORTANTE: Uso Dinamico de Agentes

**INSTRUCCION CRITICA:** Cuando trabajes en este repositorio o cualquier proyecto donde estos agentes esten instalados, DEBES usar los agentes de forma PROACTIVA y AUTOMATICA. No esperes a que el usuario te pida usar un agente especifico.

### Reglas de Activacion Automatica

Antes de responder cualquier pregunta o realizar cualquier tarea, evalua si alguno de los agentes puede ayudar:

| Contexto Detectado | Agente a Usar |
|-------------------|---------------|
| Archivos `.tf`, `.tfvars`, palabras como "terraform", "infrastructure" | `terraform` |
| Menciones de AWS, EC2, S3, Lambda, ECS, RDS | `aws` |
| Menciones de Azure, App Service, AKS, Azure Functions | `azure` |
| Archivos `Dockerfile`, `docker-compose.yml`, containers | `docker` |
| Archivos YAML de K8s, Helm charts, pods, deployments | `kubernetes` |
| Archivos `.github/workflows/`, pipelines, CI/CD | `cicd` |
| Queries SQL, migraciones, schemas de base de datos | `databases` |
| Metricas, alertas, Prometheus, Grafana, logs | `monitoring` |
| VPCs, subnets, security groups, networking | `networking` |
| Code review, patrones de diseno, tests, refactoring | `programming` |
| React, Next.js, APIs, frontend/backend web | `web` |
| **React Native, Expo, Flutter, iOS, Android, MVP mobile, apps** | `mobile` |
| Costos cloud, billing, optimization, rightsizing | `finops` |
| Vulnerabilidades, CVEs, secrets, seguridad | `security` |
| Diseno de sistemas, arquitectura, trade-offs | `architecture` |

### Como Invocar Agentes

Usa el Task tool con `subagent_type` para invocar el agente apropiado:

```
Task(subagent_type="mobile", prompt="Ayuda con...", description="Mobile dev task")
Task(subagent_type="web", prompt="Implementa...", description="Web dev task")
Task(subagent_type="terraform", prompt="Crea...", description="IaC task")
```

### Ejemplo de Uso Automatico

Si el usuario dice: "Estoy creando un MVP de una app mobile con React Native"

1. Detectar keywords: "MVP", "app", "mobile", "React Native"
2. Activar automaticamente: `mobile` (principal), `programming` (co-activacion), `web` (para backend)
3. Usar el agente `mobile` para tareas de frontend mobile
4. Usar `programming` para patrones y testing
5. Usar `databases` si hay persistencia de datos

**NO ESPERAR** a que el usuario pida explicitamente usar un agente. Detectar el contexto y usarlo proactivamente.

## Agentes Disponibles

Este repositorio contiene **15 subagentes especializados** que Claude Code usara automaticamente segun el contexto:

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
| `kubernetes` | K8s, Helm, Kustomize, operators, troubleshooting |
| `monitoring` | Prometheus, Grafana, CloudWatch, SLOs, alerting |
| `docker` | Dockerfiles, multi-stage builds, Compose, security |
| `finops` | Cost optimization, rightsizing, Savings Plans, budgets |
| `security` | CVE scanning, OWASP, secrets detection, IAM |
| `mobile` | **NUEVO** React Native, Expo, Flutter, iOS, Android, MVPs |

## Cognitive Features (v2.0)

Este repositorio incluye funcionalidades cognitivas avanzadas inspiradas en [claude-cognitive](https://github.com/GMaN1911/claude-cognitive):

### Context Router

Activacion inteligente de agentes basada en keywords:

- **HOT** (score > 0.8): Inyeccion completa del agente
- **WARM** (score 0.25-0.8): Solo headers del agente
- **COLD** (score < 0.25): Agente no inyectado

Caracteristicas:
- Activacion por palabras clave (configurado en `keywords.json`)
- Co-activacion de agentes relacionados
- Decay gradual de agentes no mencionados
- Archivos fijados (pinned) que siempre estan disponibles

### Pool Coordinator

Coordinacion entre multiples instancias de Claude Code:

- Detecta automaticamente tareas completadas y bloqueadores
- Comparte estado entre instancias (A, B, C, etc.)
- Evita trabajo duplicado en equipos

Para usar multiples instancias:
```bash
# Terminal 1
export CLAUDE_INSTANCE=A
claude

# Terminal 2
export CLAUDE_INSTANCE=B
claude
```

## Como Usar

Los agentes se activan **automaticamente** basandose en el contexto de la conversacion. Ver la seccion "Uso Dinamico de Agentes" arriba para entender como funciona.

### Ver agentes disponibles
Ejecuta `/agents` en Claude Code para ver y gestionar los agentes.

## Comandos (Skills)

Este repositorio incluye **8 comandos** listos para usar:

| Comando | Descripcion | Uso |
|---------|-------------|-----|
| `/commit` | Genera commit message con Conventional Commits | `git add . && /commit` |
| `/pr` | Crea PR con descripcion auto-generada | `/pr` |
| `/review` | Code review de cambios staged o archivos | `/review` o `/review src/file.ts` |
| `/test` | Genera unit tests para codigo | `/test src/utils.ts` |
| `/explain` | Explica como funciona el codigo | `/explain src/auth/` |
| `/refactor` | Sugiere y aplica refactoring | `/refactor src/legacy.ts` |
| `/debug` | Ayuda a debuggear errores | `/debug "TypeError: cannot read..."` |
| `/doc` | Genera documentacion (README, JSDoc, docstrings) | `/doc src/api/` |

### Ejemplos de Uso

```bash
# Hacer commit con mensaje generado
git add .
/commit

# Crear PR
/pr

# Review de codigo
/review

# Generar tests
/test src/services/auth.ts

# Explicar codigo complejo
/explain src/algorithms/dijkstra.ts

# Refactorizar codigo legacy
/refactor src/old-module.js

# Debuggear un error
/debug "Cannot read property 'map' of undefined"

# Generar documentacion
/doc src/api/
```

## MCP Servers Incluidos

Este repositorio incluye configuracion de MCP servers en `.mcp.json`:

| Server | Descripcion | Requiere Config |
|--------|-------------|-----------------|
| `memory` | Memoria persistente entre sesiones | No |
| `context7` | Documentacion actualizada de librerias (agrega "use context7" al prompt) | No |
| `supabase` | Interaccion con proyectos Supabase | OAuth automatico |
| `notion` | Acceso a workspaces de Notion | `NOTION_TOKEN` env var |

### MCP Proxy (Opcional)

Para consolidar multiples MCP servers en un unico endpoint, podes usar [mcp-proxy](https://github.com/tbxark/mcp-proxy):

```bash
# Instalar
go install github.com/tbxark/mcp-proxy@latest

# Ejecutar con la config incluida
mcp-proxy -config mcp-proxy-config.json
```

Ventajas:
- Un solo endpoint para todos los MCP servers
- Autenticacion centralizada
- Filtrado de herramientas por servidor
- Logs centralizados

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

**Tip:** Si memory se vuelve muy grande y causa problemas de tokens:
```bash
# Ver estadisticas de memory
python3 .claude/scripts/memory-manager.py stats

# Limpiar memory (hace backup automatico)
python3 .claude/scripts/memory-manager.py clear

# Exportar backup
python3 .claude/scripts/memory-manager.py export backup.json
```

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

### Opcion 1: CLI Tool (Recomendado)

```bash
# Clonar repositorio
git clone <repo-url> ~/claude-agents
cd ~/claude-agents

# Instalar con symlinks (auto-actualiza con git pull)
./scripts/claude-agents-cli.sh install

# O copiar archivos (independiente del repo)
./scripts/claude-agents-cli.sh install --copy

# Ver estado de instalacion
./scripts/claude-agents-cli.sh status

# Probar configuracion
./scripts/claude-agents-cli.sh test
```

### Opcion 2: Manual

```bash
# Clonar y copiar todo
git clone <repo-url> ~/claude-agents
cp -r ~/claude-agents/.claude ~/.claude
cp ~/claude-agents/.mcp.json ~/.mcp.json

# O symlink (recomendado para actualizaciones)
git clone <repo-url> ~/claude-agents
ln -s ~/claude-agents/.claude/agents ~/.claude/agents
ln -s ~/claude-agents/.claude/scripts ~/.claude/scripts
cp ~/claude-agents/.claude/keywords.json ~/.claude/
cp ~/claude-agents/.claude/settings.json ~/.claude/
```

### Opcion 3: Proyecto especifico

```bash
# Copiar a un proyecto
cp -r ~/claude-agents/.claude /mi-proyecto/.claude
cp ~/claude-agents/.mcp.json /mi-proyecto/.mcp.json
```

**Nota:** El archivo `.mcp.json` debe estar en la raiz del proyecto para que Claude Code lo detecte.

## Estructura del Repositorio

```
claude-agents/
├── CLAUDE.md                 # Este archivo
├── .mcp.json                 # Configuracion de MCP servers
├── mcp-proxy-config.json     # Config para mcp-proxy (opcional)
├── .claude/
│   ├── agents/               # 15 subagentes nativos
│   │   ├── architecture.md
│   │   ├── terraform.md
│   │   ├── aws.md
│   │   ├── azure.md
│   │   ├── networking.md
│   │   ├── databases.md
│   │   ├── programming.md
│   │   ├── web.md
│   │   ├── mobile.md         # NUEVO - React Native, Expo, Flutter
│   │   ├── cicd.md
│   │   ├── kubernetes.md
│   │   ├── monitoring.md
│   │   ├── docker.md
│   │   ├── finops.md
│   │   └── security.md
│   ├── commands/             # Comandos/Skills
│   │   ├── commit.md         # /commit - Conventional commits
│   │   ├── pr.md             # /pr - Create pull request
│   │   ├── review.md         # /review - Code review
│   │   ├── test.md           # /test - Generate tests
│   │   ├── explain.md        # /explain - Explain code
│   │   ├── refactor.md       # /refactor - Suggest refactoring
│   │   ├── debug.md          # /debug - Debug errors
│   │   └── doc.md            # /doc - Generate docs
│   ├── scripts/              # Scripts utilitarios
│   │   ├── context-router.py # Activacion inteligente
│   │   ├── pool-loader.py    # Carga estado de pool
│   │   ├── pool-extractor.py # Extrae senales de completado
│   │   ├── pool-query.py     # CLI para consultar pool
│   │   └── memory-manager.py # Gestionar MCP memory
│   ├── pool/                 # Estado de coordinacion
│   ├── keywords.json         # Config de palabras clave
│   └── settings.json         # Config de hooks
├── scripts/
│   ├── install.sh            # Instalador legacy
│   └── claude-agents-cli.sh  # CLI completo (v2)
├── templates/                # Templates reutilizables
└── agents/                   # Versiones originales (referencia)
```

## Hooks de Claude Code

Los hooks permiten ejecutar comandos automaticamente en diferentes momentos del ciclo de vida de Claude Code.

### Hooks Cognitivos Incluidos

Este repositorio configura automaticamente los siguientes hooks en `.claude/settings.json`:

| Hook | Script | Funcion |
|------|--------|---------|
| `UserPromptSubmit` | `context-router.py` | Activa agentes segun keywords |
| `SessionStart` | `pool-loader.py` | Carga estado de otras instancias |
| `Stop` | `pool-extractor.py` | Guarda senales de completado |

### Tipos de Hooks Disponibles

| Hook | Cuando se ejecuta |
|------|-------------------|
| `PreToolUse` | Antes de ejecutar una herramienta (puede bloquearla) |
| `PostToolUse` | Despues de que una herramienta se ejecuta correctamente |
| `UserPromptSubmit` | Cuando el usuario envia un prompt |
| `Notification` | Cuando Claude Code envia notificaciones |
| `Stop` | Cuando Claude Code termina de responder |
| `SessionStart` | Al iniciar una sesion |
| `SessionEnd` | Cuando termina una sesion |

### Configuracion de Hooks

Los hooks se configuran en `.claude/settings.json` o `~/.claude/settings.json`:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/scripts/context-router.py",
            "timeout": 5000
          }
        ]
      }
    ]
  }
}
```

### Ejemplos Utiles

**Proteger archivos sensibles:**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 -c \"import json,sys; d=json.load(sys.stdin); p=d.get('tool_input',{}).get('file_path',''); sys.exit(2 if '.env' in p or 'secrets' in p else 0)\""
          }
        ]
      }
    ]
  }
}
```

**Auto-format TypeScript:**
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_input.file_path' | xargs -I {} sh -c 'echo \"{}\" | grep -q \"\\.ts$\" && npx prettier --write \"{}\"'"
          }
        ]
      }
    ]
  }
}
```

### Codigos de Salida

- `0`: Exito, continua normalmente
- `2`: Bloquea la accion (stderr se muestra a Claude)
- Otros: Error no bloqueante

Para mas info sobre hooks, ejecuta `/hooks` en Claude Code.

## Variables de Entorno

| Variable | Descripcion | Requerido |
|----------|-------------|-----------|
| `CLAUDE_INSTANCE` | ID de instancia (A, B, C...) para coordinacion | No |
| `NOTION_TOKEN` | Token de integracion de Notion | Para MCP Notion |
| `GITHUB_TOKEN` | Personal Access Token de GitHub | Para MCP GitHub |
| `MCP_PROXY_TOKEN` | Token para mcp-proxy | Para mcp-proxy |
| `MAX_MCP_OUTPUT_TOKENS` | Limite de tokens para MCP (default: 25000) | No |

**Tip:** Si memory o MCP servers causan problemas de tokens:
```bash
export MAX_MCP_OUTPUT_TOKENS=250000
```

## Sincronizacion entre Ordenadores

Para mantener sincronizada tu configuracion entre multiples ordenadores:

```bash
# En PC 1: Clonar e instalar
git clone <repo-url> ~/claude-agents
cd ~/claude-agents
./scripts/claude-agents-cli.sh install

# En PC 2: Clonar e instalar igual
git clone <repo-url> ~/claude-agents
cd ~/claude-agents
./scripts/claude-agents-cli.sh install

# Cuando actualices el repo:
cd ~/claude-agents
git pull
./scripts/claude-agents-cli.sh sync  # Si usaste --copy
# (Si usaste symlinks, se actualiza automaticamente)
```

## Referencias

Este proyecto incorpora ideas y patrones de:
- [claude-cognitive](https://github.com/GMaN1911/claude-cognitive) - Context Router y Pool Coordinator
- [claude-code-templates](https://github.com/davila7/claude-code-templates) - Inspiracion para skills/commands
- [mcp-proxy](https://github.com/tbxark/mcp-proxy) - Agregacion de MCP servers
- [gen-ai-cve-patching](https://github.com/aws-samples/gen-ai-cve-patching) - Patrones para security agent
- [llm-engineer-toolkit](https://github.com/KalyanKS-NLP/llm-engineer-toolkit) - Catalogo de herramientas LLM
- [rag-zero-to-hero](https://github.com/KalyanKS-NLP/rag-zero-to-hero-guide) - Patrones de RAG y evaluacion

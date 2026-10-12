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

Usa el tool `Agent` con `subagent_type` para delegar en el agente apropiado:

```
Agent(subagent_type="mobile", prompt="Ayuda con...", description="Mobile dev task")
Agent(subagent_type="web", prompt="Implementa...", description="Web dev task")
Agent(subagent_type="terraform", prompt="Crea...", description="IaC task")
```

### Trabajo en Paralelo (orquestacion)

La instancia principal es la que reparte trabajo. Reglas:

- **Independiente → en paralelo.** Si dos tareas no dependen una de otra, lanza los
  `Agent` en el MISMO mensaje (varios tool calls juntos): corren concurrentemente y
  seguis trabajando mientras tanto; los reportes llegan como notificacion.
- **Dependiente → secuencial.** Si B necesita el resultado de A, espera A.
- **`subagent_type="fork"`** cuando el agente necesita TODO tu contexto actual (misma
  conversacion, mismo modelo). Agente especializado (`terraform`, `web`...) cuando lo que
  importa es su prompt experto y un contexto limpio.
- **`isolation="worktree"`** cuando dos agentes van a EDITAR el mismo repo a la vez: cada
  uno trabaja en su propio git worktree y no se pisan.
- **Investigar vs. editar.** Para auditorias, busquedas amplias o comparar archivos, pedi
  un reporte (el agente lee, vos decidis). Para implementar, dale una tarea cerrada y
  verificable (tests, `terraform plan`, lint).
- **`Workflow`** para orquestaciones grandes (pipeline review → verify, fan-out de N
  agentes): solo si el usuario lo pide ("use a workflow" / "ultracode").
- Nunca hagas vos mismo lo que ya delegaste; espera el resultado y relata lo que importa.

Ejemplo: "auditar el repo" → 2 `Agent(programming)` en paralelo (contenido e instalador),
mientras la instancia principal prueba el router a mano. Ver `agents-hub` para el detalle.

### Ejemplo de Uso Automatico

Si el usuario dice: "Estoy creando un MVP de una app mobile con React Native"

1. Detectar keywords: "MVP", "app", "mobile", "React Native"
2. Activar automaticamente: `mobile` (principal), `programming` (co-activacion), `web` (para backend)
3. Usar el agente `mobile` para tareas de frontend mobile
4. Usar `programming` para patrones y testing
5. Usar `databases` si hay persistencia de datos

**NO ESPERAR** a que el usuario pida explicitamente usar un agente. Detectar el contexto y usarlo proactivamente.

## Agentes Disponibles

Este repositorio contiene **15 subagentes especializados** que Claude Code usara automaticamente segun el contexto.
Modelo: `architecture` y `security` corren con `opus` (razonamiento, poco volumen); el resto con `sonnet`
(generacion de codigo/config, donde el costo por token pesa mas).

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

Sugerencia de agentes basada en keywords, en UNA linea por prompt:

```
[context-router] Suggested agents: terraform, aws · related: networking · skill: infrastructure-as-code
```

- **HOT** (score >= 0.8): aparece en `Suggested agents` → delegar a ese agente
- **WARM** (score 0.25-0.8): aparece en `related` → tenerlo en cuenta
- **COLD** (score < 0.25): no aparece

No inyecta el contenido de los agentes (cada subagente ya recibe su propio prompt al
invocarlo). Config en `~/.claude/keywords.json` (keywords en ingles y espanol, sin
acentos), estado en `~/.claude/router/<proyecto>.json`, historial opt-in. `"output": "full"`
restaura la inyeccion de contenido de la v2.

### Pool Coordinator

Coordinacion entre multiples **terminales** de Claude Code abiertas a la vez (vos con dos
sesiones, o un equipo). NO es lo que lanza agentes en paralelo: eso es el tool `Agent`
(ver "Trabajo en Paralelo").

- **Opt-in:** los hooks salen al instante si `CLAUDE_INSTANCE` no esta definida
- Comparte estado entre instancias (A, B, C, etc.) via `.claude/pool/instance_state.jsonl`
- Evita trabajo duplicado en equipos
- El hook `SessionStart` inyecta la actividad reciente de otras instancias al contexto
- El hook `Stop` extrae los bloques `pool` explicitos de la ultima respuesta

Para usar multiples instancias:
```bash
# Terminal 1
export CLAUDE_INSTANCE=A
claude

# Terminal 2
export CLAUDE_INSTANCE=B
claude
```

**INSTRUCCION (protocolo pool):** Si la variable `CLAUDE_INSTANCE` esta definida, al terminar
una tarea significativa (feature completada, blocker encontrado, decision que afecta a otros)
DEBES cerrar tu respuesta con un bloque `pool` para que las demas instancias lo vean:

````markdown
```pool
INSTANCE: A
ACTION: completed | blocked | in_progress | signaling
TOPIC: titulo corto de la tarea
SUMMARY: que se hizo o que bloquea, en una linea
AFFECTS: archivos o areas afectadas
BLOCKS: que trabajo puede continuar o queda bloqueado (opcional)
```
````

Solo los bloques explicitos se guardan (el hook no adivina con regex). No emitas bloques
para respuestas conversacionales o cambios triviales.

## Como Usar

Los agentes se activan **automaticamente** basandose en el contexto de la conversacion. Ver la seccion "Uso Dinamico de Agentes" arriba para entender como funciona.

### Ver agentes disponibles
Ejecuta `/agents` en Claude Code para ver y gestionar los agentes.

## Workflow Skills

Este repositorio incluye una **skill guia** (`agents-hub`), **3 skills de dia a dia** (sesion y
git, transversales a cualquier proyecto) y **5 workflow skills** que orquestan multiples agentes:

| Skill | Agentes | Descripcion |
|-------|---------|-------------|
| `agents-hub` | todos | Guia de routing: punto de entrada para elegir e invocar el agente correcto de forma proactiva (usar al inicio de cualquier tarea) |
| `deploy-pipeline` | cicd, docker, aws/azure, networking, security | Pipeline CI/CD completo: Dockerfile + workflow + cloud deploy |
| `observability-setup` | monitoring, aws/azure, kubernetes | Stack de monitoring: metricas + logging + alerting + dashboards |
| `infrastructure-as-code` | terraform, networking, aws/azure, security, architecture | IaC completo: VPC + modulos Terraform + multi-ambiente |
| `security-hardening` | security, docker, cicd, aws/azure | Audit de seguridad: dependencias + secrets + containers + pipeline |
| `full-stack-scaffold` | architecture, web/mobile, databases, programming, docker | Scaffolding: API + DB + frontend + Docker Compose + tests |
| `session-resume` | - | "Por donde nos quedamos": reconstruye contexto desde el vault de Obsidian de ESTA maquina (`~/.claude/vault.json`), git, PRs abiertos y archivos de plan |
| `session-save` | - | "Anota todo en el vault": nota de sesion + estado del proyecto + indice, segun las rules del vault; muestra el borrador antes de escribir |
| `git-flow` | - | "Commitealo y abri el PR": branch desde la base correcta, conventional commits sin Co-Authored-By, PR en el idioma del repo, confirmacion antes de push/merge, verificacion post-merge |

**Vaults por maquina.** Las skills de sesion leen `~/.claude/vault.json` (ruta del vault, MCP
a usar, carpeta de proyectos, nota indice, nota de rules). Es local y nunca se commitea: cada
PC (personal, trabajo) tiene su propio vault y la misma skill. Si falta, la skill pregunta una
vez y lo crea. El hook `session-health.py` avisa al arrancar si el vault no esta accesible.

Los skills se instalan automaticamente con `claude-agents-cli.sh install` en `.claude/skills/`.

### Integracion con Superpowers Plugin

Los workflow skills complementan el plugin superpowers:

```
superpowers:brainstorming → Refinar requerimientos
    ↓
superpowers:writing-plans → Crear plan de implementacion
    ↓
Workflow Skill (deploy-pipeline, etc.) → Ejecutar el trabajo
    ↓
superpowers:verification-before-completion → Validar resultados
```

## Comandos (Slash Commands)

Este repositorio **no instala slash commands**. Tenia 8 (`/commit`, `/pr`, `/review`,
`/test`, `/explain`, `/refactor`, `/debug`, `/doc`) y en 10 meses de historial no se usaron
ni una vez: lo mismo se pide en lenguaje natural ("commitealo y abri el PR") y los agentes
y skills lo resuelven. Ademas `/review` y `/debug` pisaban comandos nativos de Claude Code.
Los flujos que si se repiten viven como skills (ver "Workflow Skills"). Jarvis, el asistente
por voz, vive en su propio repo: https://github.com/Lkodestyle/jarvis

## Plugins Recomendados

### Esenciales

| Plugin | Descripcion |
|--------|-------------|
| `superpowers` | 14 skills de proceso: brainstorming, TDD, debugging, plans, code review, agents paralelos |
| `skill-creator` | Crear y testear skills custom con framework de evaluacion |
| `context7` | Documentacion contextual de librerias y frameworks via MCP |

### Opcionales (segun stack)

| Plugin | Caso de uso |
|--------|-------------|
| `github` / `gitlab` | Gestion de repos, PRs, issues |
| `playwright` | Testing E2E de apps web |
| `supabase` | Backend-as-a-service |
| `slack` | Notificaciones y comunicacion |
| `firebase` | Apps mobile/web con Firebase |
| `linear` | Issue tracking |

Para instalar: `claude config set enabledPlugins.superpowers@claude-plugins-official true`
Para ver todos: `./scripts/claude-agents-cli.sh plugins`

## MCP Servers Incluidos

Este repositorio incluye configuracion de MCP servers en `.mcp.json`:

| Server | Descripcion | Requiere Config |
|--------|-------------|-----------------|
| `memory` | Memoria persistente entre sesiones | No |
| `context7` | Documentacion actualizada de librerias (agrega "use context7" al prompt) | No |
| `supabase` | Interaccion con proyectos Supabase | OAuth automatico |
| `notion` | Acceso a workspaces de Notion | `NOTION_TOKEN` env var |
| `obsidian` | Lectura/escritura/busqueda en vault de Obsidian via Local REST API | `OBSIDIAN_API_KEY` env var |

### Obsidian Vault

Integra tu vault de Obsidian via el plugin **Local REST API** (nativo en Obsidian):

1. Instalar el plugin en Obsidian: Settings → Community Plugins → buscar "Local REST API"
2. Activarlo y en Settings → Local REST API → habilitar **"Enable HTTP server"**
3. Copiar la **API key** que aparece en la misma pantalla
4. Configurar la variable de entorno:

```bash
# Agregar a ~/.bashrc, ~/.zshrc, o variables de entorno de Windows
export OBSIDIAN_API_KEY="tu_api_key_aqui"
```

El servidor MCP corre en `http://127.0.0.1:27123/mcp/` mientras Obsidian esté abierto.

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

### Opcion 1: Setup Completo (Recomendado)

```bash
# Clonar repositorio
git clone <repo-url> ~/claude-agents
cd ~/claude-agents

# Setup completo interactivo (agentes + MCP + Obsidian + plugins)
./scripts/setup.sh

# Ver estado de instalacion
./scripts/claude-agents-cli.sh status
```

### Opcion 2: Solo agentes/skills

```bash
# Instalar con symlinks (auto-actualiza con git pull)
./scripts/claude-agents-cli.sh install

# O copiar archivos (independiente del repo)
./scripts/claude-agents-cli.sh install --copy

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
│   │   ├── commit.md         # /commit - Conventional commits
│   │   ├── pr.md             # /pr - Create pull request
│   │   ├── review-changes.md # /review-changes - Code review
│   │   ├── test.md           # /test - Generate tests
│   │   ├── explain.md        # /explain - Explain code
│   │   ├── refactor.md       # /refactor - Suggest refactoring
│   │   ├── diagnose.md       # /diagnose - Debug errors
│   │   └── doc.md            # /doc - Generate docs
│   ├── skills/               # Skills (guia + workflows)
│   │   ├── agents-hub/       # Guia de routing de agentes (entry point)
│   │   ├── deploy-pipeline/  # CI/CD + Docker + Cloud
│   │   ├── observability-setup/ # Monitoring + Logging + Alerting
│   │   ├── infrastructure-as-code/ # Terraform + Networking + Cloud
│   │   ├── security-hardening/    # Security audit + fix + compliance
│   │   ├── full-stack-scaffold/   # Project scaffolding completo
│   │   ├── session-resume/   # Retomar: vault + git + PRs -> donde quedamos
│   │   ├── session-save/     # Cerrar: nota de sesion + estado + indice en el vault
│   │   └── git-flow/         # Commits + PR + merge con las convenciones del repo
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
│   ├── claude-agents-cli.sh  # CLI completo (v3)
│   ├── setup-mcp.sh          # Setup de MCP servers
│   └── setup-obsidian.sh     # Setup de Obsidian vault MCP
└── templates/                # Guias de referencia (README por tipo: ecs, gh-workflow, gitlab-ci, tf-module)
```

## Hooks de Claude Code

Los hooks permiten ejecutar comandos automaticamente en diferentes momentos del ciclo de vida de Claude Code.

### Hooks Cognitivos Incluidos

Los hooks se definen una sola vez en `scripts/hooks.json` y el instalador los
**mergea** en tu `~/.claude/settings.json` global (sin pisar model, theme ni hooks
de otras herramientas). El `.claude/settings.json` del repo NO declara hooks a
proposito: si proyecto y global tienen el mismo hook, corre dos veces por evento.

| Hook | Script | Funcion |
|------|--------|---------|
| `UserPromptSubmit` | `context-router.py` | Sugiere agentes segun keywords (una linea) |
| `SessionStart` | `session-health.py` | Avisa si Obsidian/vault/hooks/agentes tienen problemas |
| `SessionStart` | `pool-loader.py` | Carga estado de otras instancias (solo con CLAUDE_INSTANCE) |
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

Los hooks se configuran en `.claude/settings.json` (proyecto) o `~/.claude/settings.json` (global).

Para hooks **de proyecto**, usa `${CLAUDE_PROJECT_DIR}` para que la ruta resuelva sin
depender del directorio actual:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "python3 \"${CLAUDE_PROJECT_DIR}/.claude/scripts/context-router.py\"",
            "timeout": 10
          }
        ]
      }
    ]
  }
}
```

**Nota:** `timeout` va en **segundos** (no milisegundos). El campo `matcher` solo aplica a
hooks de tools (`PreToolUse`/`PostToolUse`); en eventos como `UserPromptSubmit`, `Stop` o
`SessionStart` se omite.

Para la instalacion **global**, el instalador (`claude-agents-cli.sh install`) escribe los
hooks con la forma portable `$HOME/.claude/scripts/<script>.py` y el binario de Python
detectado (`python3`/`python`/`py -3`), por ejemplo:

```json
"command": "python3 \"$HOME/.claude/scripts/context-router.py\""
```

Agentes, scripts y skills se instalan **entrada por entrada** (un symlink o
copia por archivo/carpeta), asi que conviven con skills de plugins o commands propios que
tengas en `~/.claude`. `claude-agents-cli.sh status` muestra el modo de instalacion y cuantas
veces correria cada hook.

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
| `OBSIDIAN_API_KEY` | API key del plugin Local REST API de Obsidian | Para MCP Obsidian |
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

# Claude DevOps Agents Hub

Stack portable de Claude Code (agentes, skills, hooks) que Luke usa a diario en varias PCs.
Este archivo contiene SOLO reglas de comportamiento. Instalacion, hooks, MCP, variables de
entorno y estructura del repo estan en `README.md`; migraciones en `UPGRADE.md`.

## Regla 1: usar agentes de forma proactiva

Antes de responder o editar, detectar el contexto y delegar con el tool `Agent`. No esperar a
que el usuario nombre un agente. El hook del router imprime una pista por prompt
(`[context-router] Suggested agents: ...`); la skill `agents-hub` tiene el detalle.

| Contexto | Agente |
|---|---|
| `.tf`, terraform, terragrunt, IaC | `terraform` |
| AWS (EC2, ECS, Lambda, S3, RDS...) / Azure (App Service, AKS, Functions...) | `aws` / `azure` |
| Dockerfile, compose, contenedores | `docker` |
| K8s, Helm, pods, deployments | `kubernetes` |
| pipelines, GitHub Actions, deploy, release | `cicd` |
| SQL, migraciones, schemas, Redis, Mongo | `databases` |
| metricas, alertas, Prometheus, Grafana, logs | `monitoring` |
| VPC, subnets, security groups, DNS, VPN | `networking` |
| code review, patrones, tests, refactor | `programming` |
| React, Next.js, Node, FastAPI, APIs web | `web` |
| React Native, Expo, Flutter, iOS, Android | `mobile` |
| costos cloud, rightsizing, budgets | `finops` |
| CVEs, secrets, OWASP, IAM, hardening | `security` |
| diseno de sistemas, trade-offs, ADRs | `architecture` |

```
Agent(subagent_type="terraform", prompt="...", description="IaC task")
```

Modelos: `architecture` y `security` corren en `opus` (razonamiento); el resto en `sonnet`.

## Regla 2: trabajo en paralelo

La instancia principal reparte el trabajo:

- **Independiente → paralelo**: varios `Agent` en el MISMO mensaje; corren en background y
  los reportes llegan como notificacion. Mientras tanto, seguir trabajando.
- **Dependiente → secuencial**: si B necesita el resultado de A, esperar A.
- `subagent_type="fork"` cuando el agente necesita todo el contexto actual; agente
  especializado cuando lo que importa es su prompt experto y un contexto limpio.
- `isolation="worktree"` cuando dos agentes editan el mismo repo a la vez.
- Investigar/auditar → pedir un reporte. Implementar → tarea cerrada y verificable.
- `Workflow` (ultracode) solo si el usuario lo pide.
- Nunca repetir uno mismo lo que ya se delego; relatar la conclusion, no el transcript.
- Varias terminales de Claude a la vez: cross-session messaging nativo (`ListAgents` /
  `SendMessage`). El Pool Coordinator (hooks, opt-in con `CLAUDE_INSTANCE`) hace lo mismo
  por archivo; si esta definida, cerrar las tareas significativas con un bloque `pool`:

````markdown
```pool
INSTANCE: A
ACTION: completed | blocked | in_progress | signaling
TOPIC: titulo corto
SUMMARY: una linea
AFFECTS: archivos o areas
```
````

## Skills

| Skill | Cuando |
|---|---|
| `agents-hub` | Al inicio de cualquier tarea tecnica: routing y orquestacion |
| `session-resume` | "por donde nos quedamos": vault + git + PRs → estado y siguiente paso |
| `session-save` | "anota todo en el vault": nota de sesion + nota del proyecto, validada y pusheada |
| `git-flow` | "commitealo y abri el PR": branch, conventional commits, PR, merge con confirmacion |
| `deploy-pipeline` | CI/CD + Docker + deploy a cloud |
| `observability-setup` | metricas + logs + alertas + dashboards |
| `infrastructure-as-code` | VPC + modulos Terraform + multi-ambiente |
| `security-hardening` | audit: dependencias + secrets + containers + pipeline |
| `full-stack-scaffold` | proyecto nuevo: API + DB + frontend + Compose + tests |

Flujos repetidos nuevos se hacen como skill (disparan desde lenguaje natural), nunca como
slash command: los 8 que existian tuvieron 0 usos en 10 meses y se eliminaron.

Con el plugin `superpowers` (habilitado): `brainstorming` → `writing-plans` → agentes o
workflow skill → `verification-before-completion`. `test-driven-development` y
`systematic-debugging` cuando aplica. Review de codigo: `/code-review` nativo (y
`/simplify`, `/security-review`); no hay comandos propios para eso.

## Vault y sesiones (varias maquinas)

Las skills de sesion leen `~/.claude/vault.json` (ruta del vault, MCP, carpetas, nota de
rules, template, validador, git). Es por maquina y nunca se commitea: las PCs personales usan
el brain-vault (git); la PC de trabajo tiene su propio vault local. Nunca mezclar vaults ni
copiar contenido del trabajo a memoria compartida. Las reglas del vault (`CLAUDE.md` del
vault) mandan sobre las skills.

## Hooks y settings (lo que hay que saber al trabajar aca)

- Los hooks y el statusline viven UNA sola vez, en el `~/.claude/settings.json` global,
  instalados desde `scripts/hooks.json` con merge. El `.claude/settings.json` del repo NO
  declara hooks: si los tuviera, correrian dos veces por evento. `claude-agents-cli.sh status`
  lo chequea.
- `timeout` de hooks es en **segundos**.
- Router v3: una linea por prompt, sin inyectar contenido; estado en `~/.claude/router/`.
- `session-health.py` avisa al arrancar si Obsidian/vault/hooks/agentes tienen problemas.
- `statusline.py` muestra contexto, cache, costo y rate limits; se instala solo si no hay otro.
- Mantener este CLAUDE.md por debajo de 200 lineas: se carga entero en cada sesion.
- No subir `MAX_MCP_OUTPUT_TOKENS` globalmente; override por herramienta si hace falta.

## Preferencias

- Idioma: espanol casual. Codigo y comentarios: ingles. Docs: segun contexto.
- Seguridad primero: nunca exponer credenciales ni datos sensibles.
- Simplicidad y DRY; todo versionable y reproducible (IaC).
- Codigo completo y funcional, con manejo de errores y valores configurables.
- Commits sin `Co-Authored-By`; `gh` para GitHub; confirmar antes de push, merge, borrar.
- `main` tiene ruleset: los cambios van por PR.
- Al editar settings de Claude Code, el auto mode bloquea la automodificacion: dejar el
  archivo listo en el scratchpad y que Luke lo aplique con `!` (o usar el instalador).

## Referencias

- Instalacion, CLI, hooks, MCP, variables, estructura, sync entre PCs: `README.md`
- Upgrade entre versiones: `UPGRADE.md`
- Inspiracion: claude-cognitive (router/pool), claude-code-templates, mcp-proxy

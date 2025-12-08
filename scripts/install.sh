#!/bin/bash
# install.sh - Instala los agentes de Claude Code en tu sistema
#
# Uso:
#   ./scripts/install.sh           # Instala con symlink (default)
#   ./scripts/install.sh --copy    # Copia los archivos
#   ./scripts/install.sh --help    # Muestra ayuda

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
AGENTS_SOURCE="$REPO_DIR/.claude/agents"
CLAUDE_HOME="${CLAUDE_HOME:-$HOME/.claude}"
AGENTS_TARGET="$CLAUDE_HOME/agents"

show_help() {
    cat << EOF
Claude DevOps Agents Installer

Uso:
    ./scripts/install.sh [opciones]

Opciones:
    --symlink   Crea un symlink a los agentes (default)
                Ventaja: Se actualizan automaticamente con git pull

    --copy      Copia los archivos de agentes
                Ventaja: Independiente del repositorio

    --uninstall Desinstala los agentes

    --help      Muestra esta ayuda

Ejemplos:
    ./scripts/install.sh              # Instala con symlink
    ./scripts/install.sh --copy       # Copia los archivos
    ./scripts/install.sh --uninstall  # Desinstala

EOF
}

check_prerequisites() {
    log_step "Verificando prerequisitos..."

    if [[ ! -d "$AGENTS_SOURCE" ]]; then
        log_error "No se encontro el directorio de agentes: $AGENTS_SOURCE"
        log_error "Asegurate de ejecutar desde el directorio del repositorio"
        exit 1
    fi

    log_info "Directorio de agentes encontrado: $AGENTS_SOURCE"
}

backup_existing() {
    if [[ -e "$AGENTS_TARGET" ]]; then
        local backup_path="${AGENTS_TARGET}.backup.$(date +%Y%m%d_%H%M%S)"
        log_warn "Ya existe $AGENTS_TARGET"
        log_info "Creando backup en: $backup_path"
        mv "$AGENTS_TARGET" "$backup_path"
    fi
}

install_symlink() {
    log_step "Instalando agentes con symlink..."

    check_prerequisites

    # Create .claude directory if not exists
    mkdir -p "$CLAUDE_HOME"

    backup_existing

    # Create symlink
    ln -s "$AGENTS_SOURCE" "$AGENTS_TARGET"

    log_info "Symlink creado: $AGENTS_TARGET -> $AGENTS_SOURCE"
    log_info ""
    log_info "Para actualizar los agentes, simplemente ejecuta:"
    log_info "  cd $REPO_DIR && git pull"
}

install_copy() {
    log_step "Instalando agentes (copiando archivos)..."

    check_prerequisites

    # Create .claude directory if not exists
    mkdir -p "$CLAUDE_HOME"

    backup_existing

    # Copy files
    cp -r "$AGENTS_SOURCE" "$AGENTS_TARGET"

    log_info "Archivos copiados a: $AGENTS_TARGET"
    log_info ""
    log_warn "Nota: Para actualizar, deberas ejecutar este script nuevamente"
}

uninstall() {
    log_step "Desinstalando agentes..."

    if [[ -e "$AGENTS_TARGET" ]]; then
        rm -rf "$AGENTS_TARGET"
        log_info "Agentes desinstalados: $AGENTS_TARGET"
    else
        log_warn "No se encontraron agentes instalados en: $AGENTS_TARGET"
    fi
}

show_installed_agents() {
    log_info ""
    log_info "Agentes instalados:"
    echo ""
    for agent in "$AGENTS_TARGET"/*.md; do
        if [[ -f "$agent" ]]; then
            local name=$(basename "$agent" .md)
            local desc=$(grep -m1 "^description:" "$agent" | sed 's/description: //' || echo "Sin descripcion")
            printf "  ${GREEN}%-15s${NC} %s\n" "$name" "$desc"
        fi
    done
    echo ""
    log_info "Usa /agents en Claude Code para ver y gestionar los agentes"
}

main() {
    echo ""
    echo "=========================================="
    echo "  Claude DevOps Agents Installer"
    echo "=========================================="
    echo ""

    local mode="symlink"

    while [[ $# -gt 0 ]]; do
        case $1 in
            --symlink)
                mode="symlink"
                shift
                ;;
            --copy)
                mode="copy"
                shift
                ;;
            --uninstall)
                mode="uninstall"
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "Opcion desconocida: $1"
                show_help
                exit 1
                ;;
        esac
    done

    case $mode in
        symlink)
            install_symlink
            show_installed_agents
            ;;
        copy)
            install_copy
            show_installed_agents
            ;;
        uninstall)
            uninstall
            ;;
    esac

    log_info ""
    log_info "Instalacion completada!"
}

main "$@"

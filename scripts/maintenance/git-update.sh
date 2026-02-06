#!/bin/bash

# ============================================================================
# LLM-Proxy Git Workflow Script (Optimiert 2026)
# ============================================================================
# Implementiert den Standard-Workflow:
# 1. Entwicklung auf 'develop' Branch
# 2. Commit bei jedem positiven Ergebnis
# 3. Merge nach 'master' wenn stabil
# 4. Push beide Branches zu GitLab
# ============================================================================
# Usage:
#   ./git-update.sh "commit message"              # Quick mode (nur commit)
#   ./git-update.sh -m "commit message"           # Standard mode (empfohlen)
#   ./git-update.sh -r "commit message"           # Release mode (mit Tag)
#   ./git-update.sh --test -m "commit message"    # Mit Tests
#   ./git-update.sh --build -m "commit message"   # Mit Build
# ============================================================================

set -e  # Bei Fehler abbrechen

# ============================================================================
# Konfiguration
# ============================================================================

PROJECT_NAME="LLM-Proxy"
DEVELOP_BRANCH="develop"
MASTER_BRANCH="master"

# Farben
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Optionen (Defaults)
MODE="interactive"  # quick, standard, release, interactive
RUN_TESTS=false
RUN_BUILD=false
AUTO_PUSH=true
COMMIT_MESSAGE=""
RELEASE_TYPE=""

# ============================================================================
# Hilfsfunktionen
# ============================================================================

print_info() {
    echo -e "${BLUE}ℹ ${1}${NC}"
}

print_success() {
    echo -e "${GREEN}✓ ${1}${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ ${1}${NC}"
}

print_error() {
    echo -e "${RED}✗ ${1}${NC}"
}

print_header() {
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [COMMIT_MESSAGE]

Modes:
  Quick Mode      $0 "commit message"                 Nur commit auf develop
  Standard Mode   $0 -m "commit message"              Commit + merge + push (empfohlen)
  Release Mode    $0 -r "commit message"              Wie Standard + Release Tag
  Interactive     $0                                   Interaktive Auswahl

Options:
  -m, --merge         Standard Mode (commit, merge master, push both)
  -r, --release       Release Mode (+ Release Tag)
  -q, --quick         Quick Mode (nur commit, kein merge)
  --test              Tests ausführen vor Commit
  --build             Build ausführen vor Commit
  --no-push           Nicht automatisch pushen
  -h, --help          Diese Hilfe anzeigen

Examples:
  $0 "feat: Add new filter feature"
  $0 -m "fix: Resolve API bug"
  $0 -r "feat: Major feature release"
  $0 --test --build -m "refactor: Optimize performance"

Commit Conventions:
  feat:      Neue Funktion
  fix:       Bug-Fix
  docs:      Dokumentation
  refactor:  Code-Umstrukturierung
  test:      Tests
  chore:     Build/Config

EOF
}

check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Kein Git Repository gefunden!"
        exit 1
    fi
}

check_project_dir() {
    if [[ ! -f "go.mod" ]]; then
        print_error "Nicht im LLM-Proxy Projekt-Verzeichnis!"
        print_info "Bitte führe das Script aus dem Projekt-Root aus"
        exit 1
    fi
}

branch_exists() {
    git rev-parse --verify "$1" >/dev/null 2>&1
}

check_uncommitted_changes() {
    if [[ -n $(git status -s) ]]; then
        return 0  # Hat Änderungen
    else
        return 1  # Keine Änderungen
    fi
}

switch_to_develop() {
    local current_branch=$(git branch --show-current)
    
    if [[ "$current_branch" != "$DEVELOP_BRANCH" ]]; then
        print_info "Wechsle zu $DEVELOP_BRANCH Branch..."
        
        if ! branch_exists "$DEVELOP_BRANCH"; then
            print_warning "Branch '$DEVELOP_BRANCH' existiert nicht"
            read -p "Von master erstellen? (j/n): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Jj]$ ]]; then
                git branch "$DEVELOP_BRANCH"
                print_success "Branch '$DEVELOP_BRANCH' erstellt"
            else
                print_error "Abgebrochen"
                exit 1
            fi
        fi
        
        git checkout "$DEVELOP_BRANCH"
        print_success "Auf Branch '$DEVELOP_BRANCH'"
    fi
}

run_tests() {
    print_info "Führe Tests aus..."
    if go test ./... -v 2>&1 | tee /tmp/llm-proxy-test.log; then
        print_success "Tests erfolgreich"
        return 0
    else
        print_warning "Tests fehlgeschlagen"
        tail -20 /tmp/llm-proxy-test.log
        read -p "Trotzdem fortfahren? (j/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Jj]$ ]]; then
            return 0
        else
            return 1
        fi
    fi
}

run_build() {
    print_info "Führe Build aus..."
    if make build >/dev/null 2>&1; then
        print_success "Build erfolgreich"
        return 0
    else
        print_error "Build fehlgeschlagen!"
        read -p "Trotzdem fortfahren? (j/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Jj]$ ]]; then
            return 0
        else
            return 1
        fi
    fi
}

get_current_version() {
    local version=$(git tag -l "v*" | sort -V | tail -n 1)
    if [[ -z "$version" ]]; then
        echo "v0.0.0"
    else
        echo "$version"
    fi
}

increment_version() {
    local version=$1
    local increment_type=$2
    
    version=${version#v}
    IFS='.' read -r -a parts <<< "$version"
    local major="${parts[0]:-0}"
    local minor="${parts[1]:-0}"
    local patch="${parts[2]:-0}"
    
    case $increment_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
    esac
    
    echo "v${major}.${minor}.${patch}"
}

create_commit() {
    local message="$1"
    
    print_info "Erstelle Commit..."
    git add -A
    git commit -m "$message"
    
    local commit_hash=$(git rev-parse --short HEAD)
    print_success "Commit erstellt: $commit_hash"
    echo ""
    git show --stat HEAD | head -15
    echo ""
    
    return 0
}

merge_to_master() {
    print_header "Merge $DEVELOP_BRANCH → $MASTER_BRANCH"
    echo ""
    
    # Zu master wechseln
    git checkout "$MASTER_BRANCH"
    
    # Pull latest (falls remote ahead)
    if git remote get-url origin >/dev/null 2>&1; then
        git pull origin "$MASTER_BRANCH" --rebase || true
    fi
    
    # Merge develop
    if git merge "$DEVELOP_BRANCH" --no-edit; then
        print_success "Merge erfolgreich"
    else
        print_error "Merge-Konflikt!"
        print_info "Bitte Konflikte manuell lösen und dann:"
        echo "  git add <resolved-files>"
        echo "  git commit --no-edit"
        echo "  git push origin develop master"
        exit 1
    fi
    
    # Zurück zu develop
    git checkout "$DEVELOP_BRANCH"
    echo ""
}

create_release_tag() {
    print_header "Release-Tag erstellen"
    echo ""
    
    local current_version=$(get_current_version)
    print_info "Aktuelle Version: $current_version"
    echo ""
    
    if [[ -z "$RELEASE_TYPE" ]]; then
        echo "  1) PATCH Release (Bugfixes: 0.0.X)"
        echo "  2) MINOR Release (Features: 0.X.0)"
        echo "  3) MAJOR Release (Breaking: X.0.0)"
        echo ""
        read -p "Release-Typ (1-3): " -n 1 -r release_choice
        echo ""
        
        case $release_choice in
            1) RELEASE_TYPE="patch" ;;
            2) RELEASE_TYPE="minor" ;;
            3) RELEASE_TYPE="major" ;;
            *) print_error "Ungültig"; return 1 ;;
        esac
    fi
    
    local new_version=$(increment_version "$current_version" "$RELEASE_TYPE")
    
    echo ""
    read -p "Release Notes für $new_version (Enter für Standard): " release_notes
    
    git checkout "$MASTER_BRANCH"
    
    if [[ -n "$release_notes" ]]; then
        git tag -a "$new_version" -m "$release_notes"
    else
        git tag -a "$new_version" -m "Release $new_version - $COMMIT_MESSAGE"
    fi
    
    git checkout "$DEVELOP_BRANCH"
    
    print_success "Release-Tag $new_version erstellt"
    echo ""
}

push_to_remote() {
    if ! git remote get-url origin >/dev/null 2>&1; then
        print_warning "Kein Remote konfiguriert"
        return 0
    fi
    
    print_header "Push zu GitLab"
    echo ""
    
    case $MODE in
        quick)
            print_info "Pushe $DEVELOP_BRANCH..."
            git push origin "$DEVELOP_BRANCH"
            print_success "$DEVELOP_BRANCH gepusht"
            ;;
            
        standard)
            print_info "Pushe $DEVELOP_BRANCH und $MASTER_BRANCH..."
            git push origin "$DEVELOP_BRANCH"
            git push origin "$MASTER_BRANCH"
            print_success "Beide Branches gepusht"
            ;;
            
        release)
            print_info "Pushe $DEVELOP_BRANCH, $MASTER_BRANCH und Tags..."
            git push origin "$DEVELOP_BRANCH"
            git push origin "$MASTER_BRANCH"
            git push origin --tags
            print_success "Branches und Tags gepusht"
            ;;
    esac
    
    echo ""
}

show_summary() {
    print_header "✨ Fertig!"
    echo ""
    
    local current_branch=$(git branch --show-current)
    local commit_hash=$(git rev-parse --short HEAD)
    local remote_url=$(git remote get-url origin 2>/dev/null || echo "nicht konfiguriert")
    
    echo -e "  Mode:          ${CYAN}$MODE${NC}"
    echo -e "  Branch:        ${GREEN}$current_branch${NC}"
    echo -e "  Commit:        ${GREEN}$COMMIT_MESSAGE${NC}"
    echo -e "  Commit Hash:   ${CYAN}$commit_hash${NC}"
    
    if [[ "$MODE" == "release" ]]; then
        local latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "keine Tags")
        echo -e "  Release Tag:   ${GREEN}$latest_tag${NC}"
    fi
    
    if [[ "$AUTO_PUSH" == true ]]; then
        echo -e "  Remote:        ${GREEN}gepusht${NC}"
    else
        echo -e "  Remote:        ${YELLOW}nicht gepusht${NC}"
    fi
    
    echo ""
    print_info "Nützliche Befehle:"
    echo "  git log --oneline --graph -10    # Historie"
    echo "  git diff HEAD~1                  # Letzter Commit"
    echo "  ./deploy.sh                      # Production Deploy"
    echo ""
}

# ============================================================================
# Argument Parsing
# ============================================================================

parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -m|--merge)
                MODE="standard"
                shift
                if [[ -n "$1" && ! "$1" =~ ^- ]]; then
                    COMMIT_MESSAGE="$1"
                    shift
                fi
                ;;
            -r|--release)
                MODE="release"
                shift
                if [[ -n "$1" && ! "$1" =~ ^- ]]; then
                    COMMIT_MESSAGE="$1"
                    shift
                fi
                ;;
            -q|--quick)
                MODE="quick"
                shift
                if [[ -n "$1" && ! "$1" =~ ^- ]]; then
                    COMMIT_MESSAGE="$1"
                    shift
                fi
                ;;
            --test)
                RUN_TESTS=true
                shift
                ;;
            --build)
                RUN_BUILD=true
                shift
                ;;
            --no-push)
                AUTO_PUSH=false
                shift
                ;;
            --patch|--minor|--major)
                RELEASE_TYPE="${1#--}"
                shift
                ;;
            -*)
                print_error "Unbekannte Option: $1"
                show_usage
                exit 1
                ;;
            *)
                if [[ -z "$COMMIT_MESSAGE" ]]; then
                    COMMIT_MESSAGE="$1"
                    if [[ "$MODE" == "interactive" ]]; then
                        MODE="quick"
                    fi
                fi
                shift
                ;;
        esac
    done
}

# ============================================================================
# Interaktiver Modus
# ============================================================================

interactive_mode() {
    print_header "🚀 $PROJECT_NAME - Git Workflow"
    echo ""
    
    local current_branch=$(git branch --show-current)
    print_info "Aktueller Branch: $current_branch"
    
    if git remote get-url origin >/dev/null 2>&1; then
        local remote_url=$(git remote get-url origin)
        print_info "Remote: $remote_url"
    fi
    
    echo ""
    print_header "Workflow-Modi"
    echo ""
    echo "  1) Quick Mode     - Nur commit auf develop"
    echo "  2) Standard Mode  - Commit + merge master + push (empfohlen)"
    echo "  3) Release Mode   - Wie Standard + Release Tag"
    echo ""
    read -p "Auswahl (1-3, Standard=2): " -n 1 -r mode_choice
    echo ""
    
    case $mode_choice in
        1) MODE="quick" ;;
        3) MODE="release" ;;
        *) MODE="standard" ;;
    esac
    
    echo ""
    
    # Commit Message
    print_info "Commit Message eingeben:"
    echo "  Beispiele:"
    echo "  - feat: Add content filtering feature"
    echo "  - fix: Resolve API authentication bug"
    echo "  - docs: Update deployment guide"
    echo ""
    read -p "Message: " COMMIT_MESSAGE
    
    if [[ -z "$COMMIT_MESSAGE" ]]; then
        print_error "Commit Message darf nicht leer sein!"
        exit 1
    fi
    
    # Optional: Tests/Build
    if [[ "$MODE" != "quick" ]]; then
        echo ""
        read -p "Tests ausführen? (j/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Jj]$ ]]; then
            RUN_TESTS=true
        fi
        
        read -p "Build ausführen? (j/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Jj]$ ]]; then
            RUN_BUILD=true
        fi
    fi
    
    echo ""
}

# ============================================================================
# Hauptlogik
# ============================================================================

main() {
    # Basis-Checks
    check_project_dir
    check_git_repo
    
    # Arguments parsen
    parse_arguments "$@"
    
    # Interaktiver Modus falls keine Argumente
    if [[ "$MODE" == "interactive" ]]; then
        interactive_mode
    fi
    
    # Commit Message Check
    if [[ -z "$COMMIT_MESSAGE" ]]; then
        print_error "Keine Commit Message angegeben!"
        echo ""
        show_usage
        exit 1
    fi
    
    # Zu develop wechseln
    switch_to_develop
    
    # Änderungen prüfen
    echo ""
    if ! check_uncommitted_changes; then
        print_warning "Keine Änderungen zum Committen"
        exit 0
    fi
    
    print_info "Änderungen:"
    echo ""
    git status -s
    echo ""
    
    # Pre-Commit Checks
    if [[ "$RUN_BUILD" == true ]]; then
        if ! run_build; then
            exit 1
        fi
    fi
    
    if [[ "$RUN_TESTS" == true ]]; then
        if ! run_tests; then
            exit 1
        fi
    fi
    
    # Commit erstellen
    print_header "Commit auf $DEVELOP_BRANCH"
    echo ""
    if ! create_commit "$COMMIT_MESSAGE"; then
        print_error "Commit fehlgeschlagen!"
        exit 1
    fi
    
    # Mode-spezifische Aktionen
    case $MODE in
        quick)
            print_success "Quick Mode: Commit erstellt auf $DEVELOP_BRANCH"
            ;;
            
        standard)
            merge_to_master
            ;;
            
        release)
            merge_to_master
            create_release_tag
            ;;
    esac
    
    # Push zu Remote
    if [[ "$AUTO_PUSH" == true ]]; then
        push_to_remote
    else
        print_warning "Auto-Push deaktiviert (--no-push)"
        print_info "Manuell pushen mit:"
        echo "  git push origin $DEVELOP_BRANCH"
        if [[ "$MODE" != "quick" ]]; then
            echo "  git push origin $MASTER_BRANCH"
        fi
        echo ""
    fi
    
    # Zusammenfassung
    show_summary
}

# ============================================================================
# Script ausführen
# ============================================================================

main "$@"

#!/bin/bash

# ============================================================================
# LLM-Proxy Git Repository Update Script mit Release-Management
# ============================================================================
# Automatisiert den Workflow:
# 1. Optional: Tests und Build ausführen
# 2. Commit Änderungen
# 3. Optional: develop -> master merge
# 4. Optional: Release-Tag erstellen
# 5. Push zu Remote (falls konfiguriert)
# ============================================================================

set -e  # Bei Fehler abbrechen

# Farben für Output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ============================================================================
# Konfiguration
# ============================================================================

PROJECT_NAME="LLM-Proxy"
DEFAULT_BRANCH="master"
DEVELOP_BRANCH="develop"

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
    echo -e "${CYAN}${1}${NC}"
}

print_separator() {
    echo -e "${BLUE}============================================================================${NC}"
}

# Prüfe ob wir in einem Git Repository sind
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Kein Git Repository gefunden!"
        exit 1
    fi
}

# Prüfe ob Remote konfiguriert ist
check_remote() {
    if git remote | grep -q "origin"; then
        return 0
    else
        return 1
    fi
}

# Prüfe ob Branch existiert
branch_exists() {
    local branch=$1
    git rev-parse --verify "$branch" >/dev/null 2>&1
}

# Erstelle develop Branch falls nicht vorhanden
setup_develop_branch() {
    if ! branch_exists "$DEVELOP_BRANCH"; then
        print_warning "Develop Branch existiert nicht"
        read -p "Develop Branch von master erstellen? (j/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Jj]$ ]]; then
            git branch "$DEVELOP_BRANCH"
            print_success "Develop Branch erstellt"
            return 0
        else
            return 1
        fi
    fi
    return 0
}

# Prüfe auf uncommitted changes
check_uncommitted_changes() {
    if [[ -n $(git status -s) ]]; then
        print_warning "Gefundene Änderungen:"
        echo ""
        git status -s
        echo ""
        return 0
    else
        print_warning "Keine Änderungen zum Committen gefunden."
        return 1
    fi
}

# Hole die aktuelle Version aus Git Tags
get_current_version() {
    local version=$(git tag -l "v*" | sort -V | tail -n 1)
    if [[ -z "$version" ]]; then
        echo "v0.0.0"
    else
        echo "$version"
    fi
}

# Erhöhe die Versionsnummer
increment_version() {
    local version=$1
    local increment_type=$2
    
    # Entferne 'v' prefix
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

# Führe Go Tests aus
run_tests() {
    print_info "Führe Go Tests aus..."
    if go test ./... -v; then
        print_success "Alle Tests erfolgreich"
        return 0
    else
        print_error "Tests fehlgeschlagen!"
        return 1
    fi
}

# Führe Go Build aus
run_build() {
    print_info "Führe Build aus..."
    if make build; then
        print_success "Build erfolgreich"
        return 0
    else
        print_error "Build fehlgeschlagen!"
        return 1
    fi
}

# Zeige Commit-Statistik
show_commit_stats() {
    local commit_hash=$1
    echo ""
    print_header "Commit Details:"
    git show --stat "$commit_hash" | head -20
    echo ""
}

# ============================================================================
# Hauptlogik
# ============================================================================

main() {
    print_separator
    print_header "  🚀 $PROJECT_NAME - Git Update Script"
    print_separator
    echo ""
    
    # 1. Basis-Prüfungen
    check_git_repo
    
    # Aktueller Branch
    current_branch=$(git branch --show-current)
    print_info "Aktueller Branch: $current_branch"
    
    # Remote Check
    has_remote=false
    if check_remote; then
        has_remote=true
        remote_url=$(git remote get-url origin)
        print_info "Remote: $remote_url"
    else
        print_warning "Kein Remote Repository konfiguriert (nur lokale Commits)"
    fi
    echo ""
    
    # 2. Workflow-Auswahl
    print_header "Workflow-Optionen:"
    echo "  1) Simple Mode  - Nur commit & push auf aktuellem Branch"
    echo "  2) Release Mode - Commit, Tests, Build, Release-Tag, Push"
    echo "  3) Full Mode    - Develop->Master Merge, Release-Tag, Push"
    echo ""
    read -p "Auswahl (1-3): " -n 1 -r workflow_mode
    echo ""
    echo ""
    
    # 3. Prüfe auf Änderungen
    if ! check_uncommitted_changes; then
        print_error "Keine Änderungen gefunden. Script beendet."
        exit 0
    fi
    
    echo ""
    read -p "Diese Änderungen committen? (j/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Jj]$ ]]; then
        print_error "Abgebrochen."
        exit 0
    fi
    
    # 4. Commit Message
    echo ""
    print_info "Commit Message eingeben:"
    echo "  Beispiele:"
    echo "  - feat: Add content filtering with bulk import"
    echo "  - fix: Resolve API key validation issue"
    echo "  - docs: Update README with filter documentation"
    echo "  - refactor: Improve filter service performance"
    echo ""
    read -p "Commit Message: " commit_msg
    
    if [[ -z "$commit_msg" ]]; then
        print_error "Commit Message darf nicht leer sein!"
        exit 1
    fi
    
    # 5. Pre-Commit Checks (für Mode 2 & 3)
    if [[ "$workflow_mode" == "2" ]] || [[ "$workflow_mode" == "3" ]]; then
        echo ""
        print_separator
        print_header "Pre-Commit Checks"
        print_separator
        echo ""
        
        # Build Test
        read -p "Build ausführen vor Commit? (j/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Jj]$ ]]; then
            if ! run_build; then
                print_error "Build fehlgeschlagen. Trotzdem fortfahren? (j/n): "
                read -p "" -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Jj]$ ]]; then
                    exit 1
                fi
            fi
        fi
        
        # Tests (optional, da viele Tests fehlschlagen können)
        read -p "Tests ausführen vor Commit? (empfohlen) (j/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Jj]$ ]]; then
            if ! run_tests; then
                print_warning "Einige Tests fehlgeschlagen."
                read -p "Trotzdem fortfahren? (j/n): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Jj]$ ]]; then
                    exit 1
                fi
            fi
        fi
        echo ""
    fi
    
    # 6. Commit erstellen
    print_separator
    print_info "Erstelle Commit..."
    print_separator
    echo ""
    
    git add -A
    git commit -m "$commit_msg"
    commit_hash=$(git rev-parse --short HEAD)
    print_success "Commit erstellt: $commit_hash"
    show_commit_stats "$commit_hash"
    
    # 7. Workflow-spezifische Aktionen
    case $workflow_mode in
        1)
            # Simple Mode - nur commit & push
            print_info "Simple Mode: Commit erstellt auf $current_branch"
            ;;
            
        2)
            # Release Mode - Tag erstellen
            echo ""
            print_separator
            print_header "Release-Tag erstellen"
            print_separator
            echo ""
            
            current_version=$(get_current_version)
            print_info "Aktuelle Version: $current_version"
            echo ""
            
            echo "  1) MAJOR Release (Breaking Changes: X.0.0)"
            echo "  2) MINOR Release (Neue Features: 0.X.0)"
            echo "  3) PATCH Release (Bugfixes: 0.0.X)"
            echo "  4) Kein Release-Tag"
            echo ""
            read -p "Auswahl (1-4): " -n 1 -r release_choice
            echo ""
            
            create_release=false
            new_version=""
            
            case $release_choice in
                1)
                    new_version=$(increment_version "$current_version" "major")
                    create_release=true
                    ;;
                2)
                    new_version=$(increment_version "$current_version" "minor")
                    create_release=true
                    ;;
                3)
                    new_version=$(increment_version "$current_version" "patch")
                    create_release=true
                    ;;
                4)
                    print_info "Kein Release-Tag"
                    ;;
            esac
            
            if [[ "$create_release" == true ]]; then
                echo ""
                read -p "Release Notes für $new_version (optional): " release_notes
                
                if [[ -n "$release_notes" ]]; then
                    git tag -a "$new_version" -m "$release_notes"
                else
                    git tag -a "$new_version" -m "Release $new_version"
                fi
                print_success "Release-Tag $new_version erstellt"
            fi
            ;;
            
        3)
            # Full Mode - Develop->Master Merge
            echo ""
            print_separator
            print_header "Develop -> Master Merge"
            print_separator
            echo ""
            
            # Setup develop branch
            if ! setup_develop_branch; then
                print_error "Develop Branch benötigt für Full Mode"
                exit 1
            fi
            
            # Stelle sicher dass wir auf develop sind
            if [[ "$current_branch" != "$DEVELOP_BRANCH" ]]; then
                print_info "Wechsle zu $DEVELOP_BRANCH Branch..."
                git checkout "$DEVELOP_BRANCH"
            fi
            
            # Merge zu master
            print_info "Merge $DEVELOP_BRANCH -> $DEFAULT_BRANCH..."
            git checkout "$DEFAULT_BRANCH"
            
            if [[ $has_remote == true ]]; then
                git pull origin "$DEFAULT_BRANCH" || true
            fi
            
            if git merge "$DEVELOP_BRANCH" --no-ff -m "Merge $DEVELOP_BRANCH into $DEFAULT_BRANCH"; then
                print_success "Merge erfolgreich"
            else
                print_error "Merge Konflikt! Bitte manuell lösen."
                exit 1
            fi
            
            # Release-Tag erstellen
            echo ""
            current_version=$(get_current_version)
            print_info "Aktuelle Version: $current_version"
            echo ""
            
            echo "  1) MAJOR Release (Breaking Changes: X.0.0)"
            echo "  2) MINOR Release (Neue Features: 0.X.0)"
            echo "  3) PATCH Release (Bugfixes: 0.0.X)"
            echo ""
            read -p "Release-Typ (1-3): " -n 1 -r release_choice
            echo ""
            
            case $release_choice in
                1) new_version=$(increment_version "$current_version" "major") ;;
                2) new_version=$(increment_version "$current_version" "minor") ;;
                3) new_version=$(increment_version "$current_version" "patch") ;;
                *) print_error "Ungültige Auswahl!"; exit 1 ;;
            esac
            
            echo ""
            read -p "Release Notes für $new_version: " release_notes
            
            if [[ -n "$release_notes" ]]; then
                git tag -a "$new_version" -m "$release_notes"
            else
                git tag -a "$new_version" -m "Release $new_version"
            fi
            print_success "Release-Tag $new_version erstellt"
            
            # Zurück zu develop
            git checkout "$DEVELOP_BRANCH"
            ;;
    esac
    
    # 8. Push zu Remote (falls vorhanden)
    if [[ $has_remote == true ]]; then
        echo ""
        print_separator
        print_header "Push zu Remote"
        print_separator
        echo ""
        
        read -p "Änderungen zu Remote pushen? (j/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Jj]$ ]]; then
            print_info "Pushe zu Remote..."
            
            case $workflow_mode in
                1)
                    git push origin "$current_branch"
                    print_success "$current_branch gepusht"
                    ;;
                2)
                    git push origin "$current_branch"
                    print_success "$current_branch gepusht"
                    
                    if [[ -n "$new_version" ]]; then
                        git push origin "$new_version"
                        print_success "Tag $new_version gepusht"
                    fi
                    ;;
                3)
                    git push origin "$DEFAULT_BRANCH"
                    print_success "$DEFAULT_BRANCH gepusht"
                    
                    git push origin "$DEVELOP_BRANCH"
                    print_success "$DEVELOP_BRANCH gepusht"
                    
                    git push origin "$new_version"
                    print_success "Tag $new_version gepusht"
                    ;;
            esac
        fi
    else
        print_warning "Kein Remote konfiguriert - keine Push-Aktion"
        echo ""
        print_info "Remote hinzufügen mit:"
        echo "  git remote add origin <repository-url>"
    fi
    
    # 9. Zusammenfassung
    echo ""
    print_separator
    print_success "✨ Repository erfolgreich aktualisiert!"
    print_separator
    echo ""
    
    print_header "Zusammenfassung:"
    echo ""
    echo -e "  Branch:        ${GREEN}$(git branch --show-current)${NC}"
    echo -e "  Commit:        ${GREEN}$commit_msg${NC}"
    echo -e "  Commit Hash:   ${CYAN}$commit_hash${NC}"
    
    if [[ -n "$new_version" ]]; then
        echo -e "  Release:       ${GREEN}$new_version${NC}"
    fi
    
    if [[ $has_remote == true ]]; then
        echo -e "  Remote:        ${GREEN}gepusht${NC}"
    else
        echo -e "  Remote:        ${YELLOW}nicht konfiguriert${NC}"
    fi
    
    echo ""
    
    # Zeige nützliche Befehle
    print_header "Nützliche Befehle:"
    echo ""
    echo "  git log --oneline -5        # Letzte 5 Commits"
    echo "  git tag -l                  # Alle Tags"
    echo "  git show $commit_hash       # Commit Details"
    if [[ -n "$new_version" ]]; then
        echo "  git show $new_version       # Release Details"
    fi
    echo ""
    
    print_success "🎉 Fertig!"
    echo ""
}

# ============================================================================
# Script ausführen
# ============================================================================

# Prüfe ob im richtigen Verzeichnis
if [[ ! -f "go.mod" ]]; then
    print_error "Nicht im LLM-Proxy Projekt-Verzeichnis!"
    print_info "Bitte führe das Script aus dem Projekt-Root aus:"
    print_info "  cd /home/krieger/Sites/golang-projekte/llm-proxy"
    print_info "  ./git-update.sh"
    exit 1
fi

main "$@"

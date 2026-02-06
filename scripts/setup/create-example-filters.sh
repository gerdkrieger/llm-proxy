#!/bin/bash

# Script zum Erstellen von Beispiel-Filtern für das LLM-Proxy System
# Verwendung: ./create-example-filters.sh

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_KEY="${ADMIN_KEY:-admin_dev_local_key_12345678901234567890123456789012}"

echo "=========================================="
echo "Content Filter Setup - Beispiele"
echo "=========================================="
echo ""

# Farben für Output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Helper Funktion zum Erstellen von Filtern
create_filter() {
    local pattern=$1
    local replacement=$2
    local description=$3
    local filter_type=$4
    local priority=$5
    
    echo -e "${YELLOW}Erstelle Filter: $pattern → $replacement${NC}"
    
    response=$(curl -s -X POST "$BASE_URL/admin/filters" \
        -H "X-Admin-API-Key: $ADMIN_KEY" \
        -H "Content-Type: application/json" \
        -d "{
            \"pattern\": \"$pattern\",
            \"replacement\": \"$replacement\",
            \"description\": \"$description\",
            \"filter_type\": \"$filter_type\",
            \"case_sensitive\": false,
            \"enabled\": true,
            \"priority\": $priority
        }")
    
    filter_id=$(echo "$response" | jq -r '.id')
    if [ "$filter_id" != "null" ] && [ -n "$filter_id" ]; then
        echo -e "${GREEN}✓ Filter erstellt mit ID: $filter_id${NC}"
        echo "$response" | jq '.'
    else
        echo "✗ Fehler beim Erstellen des Filters"
        echo "$response" | jq '.'
    fi
    echo ""
}

echo "==========================================" 
echo "1. SCHIMPFWÖRTER / OFFENSIVE SPRACHE"
echo "==========================================" 
echo ""

# Beispiel 1: Einfaches Schimpfwort
create_filter \
    "badword" \
    "[GEFILTERT]" \
    "Filtert offensive Sprache" \
    "word" \
    100

# Beispiel 2: Weiteres Schimpfwort
create_filter \
    "damn" \
    "[*]" \
    "Milde Kraftausdrücke filtern" \
    "word" \
    100

echo "==========================================" 
echo "2. VERTRAULICHE INFORMATIONEN"
echo "==========================================" 
echo ""

# Beispiel 3: Phrase Filter für vertrauliche Infos
create_filter \
    "confidential information" \
    "[VERTRAULICH_ENTFERNT]" \
    "Filtert Erwähnungen vertraulicher Informationen" \
    "phrase" \
    95

# Beispiel 4: Firmeninterne Projektnamen
create_filter \
    "Project Phoenix" \
    "[INTERNES_PROJEKT]" \
    "Schützt interne Projektnamen" \
    "phrase" \
    95

echo "==========================================" 
echo "3. PERSÖNLICHE DATEN (PII)"
echo "==========================================" 
echo ""

# Beispiel 5: Email-Adressen (Regex)
create_filter \
    "\\\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\\\.[A-Z|a-z]{2,}\\\\b" \
    "[EMAIL_ENTFERNT]" \
    "Filtert Email-Adressen aus dem Text" \
    "regex" \
    90

# Beispiel 6: Telefonnummern (Regex - Deutsches Format)
create_filter \
    "\\\\b0[0-9]{2,4}[-\\\\s]?[0-9]{3,8}\\\\b" \
    "[TELEFON_ENTFERNT]" \
    "Filtert deutsche Telefonnummern" \
    "regex" \
    90

# Beispiel 7: Kreditkartennummern (Regex - Einfaches Pattern)
create_filter \
    "\\\\b[0-9]{4}[\\\\s-]?[0-9]{4}[\\\\s-]?[0-9]{4}[\\\\s-]?[0-9]{4}\\\\b" \
    "[KREDITKARTE_ENTFERNT]" \
    "Filtert Kreditkartennummern" \
    "regex" \
    95

echo "==========================================" 
echo "4. WETTBEWERBS-INFORMATIONEN"
echo "==========================================" 
echo ""

# Beispiel 8: Konkurrent-Namen
create_filter \
    "CompetitorX" \
    "[KONKURRENT]" \
    "Filtert Erwähnungen von Konkurrenten" \
    "word" \
    80

echo "==========================================" 
echo "ZUSAMMENFASSUNG"
echo "==========================================" 
echo ""

# Alle Filter auflisten
echo -e "${YELLOW}Alle erstellten Filter:${NC}"
curl -s "$BASE_URL/admin/filters" \
    -H "X-Admin-API-Key: $ADMIN_KEY" | jq '.filters[] | {id, pattern, replacement, filter_type, priority, enabled}'

echo ""
echo -e "${GREEN}✓ Setup abgeschlossen!${NC}"
echo ""
echo "==========================================" 
echo "NÄCHSTE SCHRITTE:"
echo "==========================================" 
echo ""
echo "1. Filter testen:"
echo "   curl -X POST $BASE_URL/admin/filters/test \\"
echo "     -H 'X-Admin-API-Key: $ADMIN_KEY' \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"text\": \"Test text\", \"pattern\": \"badword\", \"replacement\": \"[FILTERED]\", \"filter_type\": \"word\"}'"
echo ""
echo "2. Filter in Chat-Request testen:"
echo "   # Erst OAuth Token holen, dann Chat Request senden"
echo "   ./test-content-filters.sh"
echo ""
echo "3. Filter verwalten:"
echo "   # Liste alle Filter:"
echo "   curl $BASE_URL/admin/filters -H 'X-Admin-API-Key: $ADMIN_KEY'"
echo ""
echo "   # Filter deaktivieren:"
echo "   curl -X PUT $BASE_URL/admin/filters/{id} \\"
echo "     -H 'X-Admin-API-Key: $ADMIN_KEY' \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"enabled\": false}'"
echo ""
echo "   # Filter löschen:"
echo "   curl -X DELETE $BASE_URL/admin/filters/{id} \\"
echo "     -H 'X-Admin-API-Key: $ADMIN_KEY'"
echo ""
echo "4. Statistiken anzeigen:"
echo "   curl $BASE_URL/admin/filters/stats -H 'X-Admin-API-Key: $ADMIN_KEY'"
echo ""

#!/bin/bash
# =============================================================================
# SCHEMA VALIDATION & REPAIR SCRIPT
# =============================================================================
# Ensures database schema matches application expectations
# Runs before migrations to fix any inconsistencies
# =============================================================================

set -e

DB_CONTAINER="llm-proxy-postgres"
DB_USER="proxy_user"
DB_NAME="llm_proxy"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Schema Validation & Repair${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to check if column exists
check_column() {
    local table=$1
    local column=$2
    docker exec $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -tAc \
        "SELECT COUNT(*) FROM information_schema.columns WHERE table_name='$table' AND column_name='$column';" 2>/dev/null || echo "0"
}

# Function to check column type
get_column_type() {
    local table=$1
    local column=$2
    docker exec $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -tAc \
        "SELECT data_type FROM information_schema.columns WHERE table_name='$table' AND column_name='$column';" 2>/dev/null || echo ""
}

# Function to add missing column
add_column() {
    local table=$1
    local column=$2
    local definition=$3
    echo -e "${YELLOW}  → Adding missing column: $table.$column${NC}"
    docker exec $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -c \
        "ALTER TABLE $table ADD COLUMN IF NOT EXISTS $column $definition;" >/dev/null 2>&1
}

# Function to fix column type
fix_column_type() {
    local table=$1
    local column=$2
    local new_type=$3
    local conversion=$4
    echo -e "${YELLOW}  → Fixing column type: $table.$column → $new_type${NC}"
    docker exec $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -c \
        "ALTER TABLE $table ALTER COLUMN $column TYPE $new_type USING $conversion;" >/dev/null 2>&1
}

# ============================================================================
# VALIDATE request_logs TABLE
# ============================================================================

echo -e "${BLUE}Validating request_logs table...${NC}"

REQUIRED_COLUMNS=(
    "id:uuid"
    "client_id:uuid"
    "request_id:character varying"
    "method:character varying"
    "path:character varying"
    "model:character varying"
    "provider:character varying"
    "prompt_tokens:integer"
    "completion_tokens:integer"
    "total_tokens:integer"
    "cost_usd:numeric"
    "duration_ms:integer"
    "status_code:integer"
    "cached:boolean"
    "ip_address:character varying"
    "user_agent:text"
    "error_message:text"
    "created_at:timestamp without time zone"
    "auth_type:character varying"
    "api_key_name:character varying"
    "was_filtered:boolean"
    "filter_reason:text"
    "request_headers:jsonb"
    "request_body:text"
    "response_headers:jsonb"
    "response_body:text"
    "response_size_bytes:bigint"
)

ERRORS=0

for col_def in "${REQUIRED_COLUMNS[@]}"; do
    IFS=':' read -r col expected_type <<< "$col_def"
    
    count=$(check_column "request_logs" "$col")
    
    if [ "$count" -eq 0 ]; then
        echo -e "${RED}  ✗ Missing column: $col${NC}"
        ERRORS=$((ERRORS + 1))
        
        # Add column with appropriate type
        case $col in
            "method") add_column "request_logs" "$col" "VARCHAR(10)" ;;
            "path") add_column "request_logs" "$col" "VARCHAR(255)" ;;
            "ip_address") add_column "request_logs" "$col" "VARCHAR(50)" ;;
            "auth_type") add_column "request_logs" "$col" "VARCHAR(20)" ;;
            "api_key_name") add_column "request_logs" "$col" "VARCHAR(100)" ;;
            "was_filtered") add_column "request_logs" "$col" "BOOLEAN DEFAULT FALSE" ;;
            "filter_reason") add_column "request_logs" "$col" "TEXT" ;;
            "request_headers") add_column "request_logs" "$col" "JSONB" ;;
            "request_body") add_column "request_logs" "$col" "TEXT" ;;
            "response_headers") add_column "request_logs" "$col" "JSONB" ;;
            "response_body") add_column "request_logs" "$col" "TEXT" ;;
            "response_size_bytes") add_column "request_logs" "$col" "BIGINT" ;;
            *) echo -e "${RED}  ! Unknown column: $col${NC}" ;;
        esac
    else
        actual_type=$(get_column_type "request_logs" "$col")
        
        # Special case: ip_address should be VARCHAR, not INET
        if [ "$col" = "ip_address" ] && [ "$actual_type" = "inet" ]; then
            echo -e "${RED}  ✗ Wrong type for $col: $actual_type (expected: character varying)${NC}"
            fix_column_type "request_logs" "$col" "VARCHAR(50)" "$col::VARCHAR"
            ERRORS=$((ERRORS + 1))
        else
            echo -e "${GREEN}  ✓ $col ($actual_type)${NC}"
        fi
    fi
done

# ============================================================================
# VALIDATE INDEXES
# ============================================================================

echo ""
echo -e "${BLUE}Validating indexes...${NC}"

REQUIRED_INDEXES=(
    "idx_request_logs_request_id"
    "idx_request_logs_client_id"
    "idx_request_logs_created_at"
    "idx_request_logs_method"
    "idx_request_logs_path"
    "idx_request_logs_model"
    "idx_request_logs_status_code"
    "idx_request_logs_cached"
    "idx_request_logs_api_key_name"
    "idx_request_logs_was_filtered"
    "idx_request_logs_auth_type"
)

for idx in "${REQUIRED_INDEXES[@]}"; do
    count=$(docker exec $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -tAc \
        "SELECT COUNT(*) FROM pg_indexes WHERE indexname='$idx';" 2>/dev/null || echo "0")
    
    if [ "$count" -eq 0 ]; then
        echo -e "${YELLOW}  → Creating missing index: $idx${NC}"
        case $idx in
            "idx_request_logs_request_id") 
                docker exec $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -c \
                    "CREATE UNIQUE INDEX IF NOT EXISTS $idx ON request_logs(request_id);" >/dev/null 2>&1
                ;;
            "idx_request_logs_created_at")
                docker exec $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -c \
                    "CREATE INDEX IF NOT EXISTS $idx ON request_logs(created_at DESC);" >/dev/null 2>&1
                ;;
            *)
                col="${idx#idx_request_logs_}"
                docker exec $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -c \
                    "CREATE INDEX IF NOT EXISTS $idx ON request_logs($col);" >/dev/null 2>&1
                ;;
        esac
    else
        echo -e "${GREEN}  ✓ $idx${NC}"
    fi
done

# ============================================================================
# SUMMARY
# ============================================================================

echo ""
if [ $ERRORS -gt 0 ]; then
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}Schema repaired: $ERRORS issue(s) fixed${NC}"
    echo -e "${YELLOW}========================================${NC}"
else
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}Schema validation: PASSED${NC}"
    echo -e "${GREEN}========================================${NC}"
fi
echo ""

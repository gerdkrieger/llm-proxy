# Scripts

This directory contains all shell scripts organized by purpose.

## Directory Structure

### 📁 setup/
Scripts for initial setup and configuration:
- `create-example-filters.sh` - Create example content filters
- `update-caddy-config.sh` - Update Caddy reverse proxy configuration

### 📁 maintenance/
Scripts for server maintenance and fixes:
- `diagnose-live.sh` - Diagnose live server issues
- `fix-live-database.sh` - Fix database schema issues
- `fix-provider-models-id-type.sh` - Fix provider model ID types
- `fix-provider-models-schema.sh` - Fix provider models schema
- `git-update.sh` - Update from git repository
- `QUICK-FIX-DOCKER.sh` - Quick Docker fixes
- `rebuild-admin-ui.sh` - Rebuild admin UI container
- `restart-server.sh` - Restart server services
- `stop-server.sh` - Stop server services

### 📁 testing/
Scripts for testing:
- `test_admin_api.sh` - Test admin API endpoints
- `test-all-filters.sh` - Test all content filters
- `test_api.sh` - Test main API endpoints
- `test-content-filters.sh` - Test content filtering functionality

### 📁 Root Level Scripts
Development and operations scripts:
- `start-all.sh` - Start all services
- `start-dev.sh` - Start development environment
- `status.sh` - Check service status
- `stop-all.sh` - Stop all services

## Usage Examples

### Start Development
```bash
./scripts/start-dev.sh
```

### Check Status
```bash
./scripts/status.sh
```

### Run Tests
```bash
./scripts/testing/test_admin_api.sh
./scripts/testing/test-content-filters.sh
```

### Maintenance
```bash
# Diagnose issues
./scripts/maintenance/diagnose-live.sh

# Restart services
./scripts/maintenance/restart-server.sh

# Fix database
./scripts/maintenance/fix-live-database.sh
```

### Setup
```bash
# Create example filters
./scripts/setup/create-example-filters.sh

# Update Caddy config
./scripts/setup/update-caddy-config.sh
```

## Script Naming Convention

- `start-*.sh` - Start services
- `stop-*.sh` - Stop services
- `test-*.sh` - Run tests
- `fix-*.sh` - Fix specific issues
- `create-*.sh` - Create/setup resources
- `update-*.sh` - Update configurations
- Other utilities - Descriptive names

## Making Scripts Executable

If a script is not executable:
```bash
chmod +x scripts/path/to/script.sh
```

## Contributing

When adding new scripts:
1. Place in appropriate subdirectory
2. Use descriptive names
3. Add usage documentation at top of script
4. Update this README
5. Make executable: `chmod +x`

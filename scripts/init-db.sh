#!/bin/bash
set -e

# Database initialization script for local development
# This runs as part of PostgreSQL container startup

echo "Initializing AI-IDP platform database..."

# Create additional databases for tenant isolation testing
# In production, these would be created dynamically
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create development tenant databases
    CREATE DATABASE tenant_demo1;
    CREATE DATABASE tenant_demo2;
    CREATE DATABASE tenant_system;
    
    -- Grant access to platform user
    GRANT ALL PRIVILEGES ON DATABASE tenant_demo1 TO platform;
    GRANT ALL PRIVILEGES ON DATABASE tenant_demo2 TO platform;
    GRANT ALL PRIVILEGES ON DATABASE tenant_system TO platform;
    
    -- Create read-only user for reporting/analytics
    CREATE USER platform_readonly WITH PASSWORD 'readonly_password';
    GRANT CONNECT ON DATABASE platform TO platform_readonly;
    GRANT USAGE ON ALL SCHEMAS IN DATABASE platform TO platform_readonly;
    GRANT SELECT ON ALL TABLES IN DATABASE platform TO platform_readonly;
    
    -- Create backup user
    CREATE USER platform_backup WITH PASSWORD 'backup_password';
    GRANT CONNECT ON DATABASE platform TO platform_backup;
    ALTER USER platform_backup WITH REPLICATION;
EOSQL

echo "Database initialization completed successfully!"

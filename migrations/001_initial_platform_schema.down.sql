-- Rollback initial platform database schema
-- Drops all tables and schemas in reverse dependency order

-- Drop triggers first
DROP TRIGGER IF EXISTS update_repositories_updated_at ON git_integration.repositories;
DROP TRIGGER IF EXISTS update_users_updated_at ON user_management.users;
DROP TRIGGER IF EXISTS update_resources_updated_at ON resource_management.resources;
DROP TRIGGER IF EXISTS update_teams_updated_at ON resource_management.teams;
DROP TRIGGER IF EXISTS update_applications_updated_at ON resource_management.applications;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON control_plane.tenants;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in dependency order (child tables first)
-- ====================================================

-- Git Integration tables
DROP TABLE IF EXISTS git_integration.webhook_events;
DROP TABLE IF EXISTS git_integration.repositories;

-- Audit System tables  
DROP TABLE IF EXISTS audit_system.audit_log;

-- User Management tables
DROP TABLE IF EXISTS user_management.tenant_memberships;
DROP TABLE IF EXISTS user_management.users;

-- API Gateway tables
DROP TABLE IF EXISTS api_gateway.rate_limits;
DROP TABLE IF EXISTS api_gateway.api_keys;

-- Resource Management tables (dependent on applications and tenants)
DROP TABLE IF EXISTS resource_management.resources;
DROP TABLE IF EXISTS resource_management.teams;
DROP TABLE IF EXISTS resource_management.applications;

-- Control Plane tables (reconciliation_operations depends on tenants)
DROP TABLE IF EXISTS control_plane.reconciliation_operations;
DROP TABLE IF EXISTS control_plane.tenants;

-- Drop schemas
-- =============
DROP SCHEMA IF EXISTS git_integration;
DROP SCHEMA IF EXISTS audit_system;
DROP SCHEMA IF EXISTS user_management;
DROP SCHEMA IF EXISTS api_gateway;
DROP SCHEMA IF EXISTS resource_management;
DROP SCHEMA IF EXISTS control_plane;

-- Drop extensions (only if we're sure no other databases need them)
-- Note: Commented out to be safe in shared environments
-- DROP EXTENSION IF EXISTS "pgcrypto";

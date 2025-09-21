-- Initial platform database schema
-- Creates component schemas and foundational tables

-- Enable UUID generation extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create component schemas for platform separation
CREATE SCHEMA IF NOT EXISTS control_plane;
CREATE SCHEMA IF NOT EXISTS api_gateway;
CREATE SCHEMA IF NOT EXISTS resource_management;  
CREATE SCHEMA IF NOT EXISTS git_integration;
CREATE SCHEMA IF NOT EXISTS user_management;
CREATE SCHEMA IF NOT EXISTS audit_system;

-- Control Plane Schema
-- ===================

-- Tenants table - core tenant management
CREATE TABLE control_plane.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    database_name VARCHAR(63) NOT NULL, -- PostgreSQL identifier limit
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    settings JSONB NOT NULL DEFAULT '{}',
    resource_limits JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT valid_status CHECK (status IN ('active', 'suspended', 'terminating', 'terminated')),
    CONSTRAINT valid_db_name CHECK (database_name ~ '^[a-z][a-z0-9_]*$')
);

-- Indexes for performance
CREATE INDEX idx_tenants_status ON control_plane.tenants(status);
CREATE INDEX idx_tenants_created ON control_plane.tenants(created_at);

-- Reconciliation operations - tracks control plane operations
CREATE TABLE control_plane.reconciliation_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES control_plane.tenants(id) ON DELETE CASCADE,
    resource_type VARCHAR(100) NOT NULL,
    resource_name VARCHAR(255) NOT NULL,
    operation_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    desired_state JSONB NOT NULL,
    current_state JSONB DEFAULT '{}',
    error_message TEXT,
    reconciler_name VARCHAR(100) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    
    CONSTRAINT valid_operation_type CHECK (operation_type IN ('create', 'update', 'delete')),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'retrying'))
);

-- Indexes for reconciliation queries
CREATE INDEX idx_reconciliation_status ON control_plane.reconciliation_operations(status);
CREATE INDEX idx_reconciliation_tenant ON control_plane.reconciliation_operations(tenant_id);
CREATE INDEX idx_reconciliation_type ON control_plane.reconciliation_operations(resource_type);
CREATE INDEX idx_reconciliation_pending ON control_plane.reconciliation_operations(status, started_at) WHERE status IN ('pending', 'retrying');

-- Resource Management Schema
-- =========================

-- Applications - core application metadata (tenant-scoped)
CREATE TABLE resource_management.applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES control_plane.tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Application specification (from YAML schema)
    team_name VARCHAR(255) NOT NULL,
    owner_email VARCHAR(255) NOT NULL,
    lifecycle VARCHAR(50) NOT NULL DEFAULT 'development',
    environment_name VARCHAR(100),
    environment_region VARCHAR(100),
    
    -- Resource quotas and limits
    resource_quota JSONB NOT NULL DEFAULT '{}',
    
    -- Compliance and governance
    compliance_settings JSONB NOT NULL DEFAULT '{}',
    
    -- Dependencies (array of application references)
    dependencies JSONB NOT NULL DEFAULT '[]',
    
    -- Observability settings
    observability_config JSONB NOT NULL DEFAULT '{}',
    
    -- Status and metadata
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    conditions JSONB NOT NULL DEFAULT '[]',
    current_resources JSONB NOT NULL DEFAULT '{}',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    
    CONSTRAINT valid_lifecycle CHECK (lifecycle IN ('development', 'staging', 'production', 'deprecated')),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'running', 'failed', 'terminating')),
    UNIQUE(tenant_id, name)
);

-- Indexes for application queries
CREATE INDEX idx_applications_tenant ON resource_management.applications(tenant_id);
CREATE INDEX idx_applications_team ON resource_management.applications(tenant_id, team_name);
CREATE INDEX idx_applications_status ON resource_management.applications(status);
CREATE INDEX idx_applications_lifecycle ON resource_management.applications(lifecycle);

-- Teams - team metadata and ownership (tenant-scoped)
CREATE TABLE resource_management.teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES control_plane.tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Team structure
    lead_email VARCHAR(255) NOT NULL,
    members JSONB NOT NULL DEFAULT '[]', -- Array of member objects
    
    -- Contact information
    contacts JSONB NOT NULL DEFAULT '{}',
    
    -- Organizational structure
    department VARCHAR(255),
    organization VARCHAR(255),
    manager_email VARCHAR(255),
    
    -- Resource ownership
    owned_applications JSONB NOT NULL DEFAULT '[]',
    owned_domains JSONB NOT NULL DEFAULT '[]',
    owned_repositories JSONB NOT NULL DEFAULT '[]',
    
    -- Policies and governance
    policies JSONB NOT NULL DEFAULT '{}',
    
    -- Budget and cost management
    budget_config JSONB NOT NULL DEFAULT '{}',
    
    -- Status tracking
    member_count INTEGER NOT NULL DEFAULT 0,
    active_applications INTEGER NOT NULL DEFAULT 0,
    monthly_spend DECIMAL(10,2) DEFAULT 0.00,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    
    UNIQUE(tenant_id, name)
);

-- Indexes for team queries
CREATE INDEX idx_teams_tenant ON resource_management.teams(tenant_id);
CREATE INDEX idx_teams_lead ON resource_management.teams(tenant_id, lead_email);
CREATE INDEX idx_teams_department ON resource_management.teams(department);

-- Resources - generic resource tracking across providers
CREATE TABLE resource_management.resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES control_plane.tenants(id) ON DELETE CASCADE,
    application_id UUID REFERENCES resource_management.applications(id) ON DELETE CASCADE,
    
    -- Resource identification
    name VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL, -- Service, Database, Repository, etc.
    api_version VARCHAR(50) NOT NULL,
    kind VARCHAR(100) NOT NULL,
    
    -- Resource specification and status
    spec JSONB NOT NULL DEFAULT '{}',
    status JSONB NOT NULL DEFAULT '{}',
    
    -- Provider information
    provider_type VARCHAR(100) NOT NULL, -- github, kubernetes, azure, etc.
    provider_resource_id VARCHAR(500), -- External resource identifier
    
    -- Ownership and metadata
    owner_team VARCHAR(255),
    labels JSONB NOT NULL DEFAULT '{}',
    annotations JSONB NOT NULL DEFAULT '{}',
    
    -- Lifecycle management
    desired_state JSONB NOT NULL DEFAULT '{}',
    current_state JSONB NOT NULL DEFAULT '{}',
    reconciliation_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    last_reconciled_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    
    CONSTRAINT valid_reconciliation_status CHECK (reconciliation_status IN ('pending', 'synced', 'error', 'unknown')),
    UNIQUE(tenant_id, name, resource_type)
);

-- Indexes for resource queries
CREATE INDEX idx_resources_tenant ON resource_management.resources(tenant_id);
CREATE INDEX idx_resources_application ON resource_management.resources(application_id);
CREATE INDEX idx_resources_type ON resource_management.resources(resource_type);
CREATE INDEX idx_resources_provider ON resource_management.resources(provider_type);
CREATE INDEX idx_resources_reconciliation ON resource_management.resources(reconciliation_status);
CREATE INDEX idx_resources_team ON resource_management.resources(owner_team);

-- API Gateway Schema
-- =================

-- API keys and authentication tokens
CREATE TABLE api_gateway.api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES control_plane.tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE, -- SHA-256 of actual key
    permissions JSONB NOT NULL DEFAULT '[]', -- Array of permission strings
    scopes JSONB NOT NULL DEFAULT '[]', -- Array of scope strings
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_api_keys_tenant ON api_gateway.api_keys(tenant_id);
CREATE INDEX idx_api_keys_active ON api_gateway.api_keys(is_active);
CREATE INDEX idx_api_keys_hash ON api_gateway.api_keys(key_hash);

-- Request rate limiting
CREATE TABLE api_gateway.rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES control_plane.tenants(id) ON DELETE CASCADE,
    identifier VARCHAR(255) NOT NULL, -- IP, user ID, API key, etc.
    identifier_type VARCHAR(50) NOT NULL, -- ip, user, api_key
    endpoint_pattern VARCHAR(500) NOT NULL,
    request_count INTEGER NOT NULL DEFAULT 0,
    window_start TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    window_duration_seconds INTEGER NOT NULL DEFAULT 3600, -- 1 hour default
    limit_per_window INTEGER NOT NULL DEFAULT 1000,
    
    CONSTRAINT valid_identifier_type CHECK (identifier_type IN ('ip', 'user', 'api_key', 'tenant'))
);

CREATE INDEX idx_rate_limits_identifier ON api_gateway.rate_limits(identifier, endpoint_pattern);
CREATE INDEX idx_rate_limits_window ON api_gateway.rate_limits(window_start);

-- User Management Schema
-- =====================

-- Users - platform user accounts (cross-tenant)
CREATE TABLE user_management.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100),
    full_name VARCHAR(255),
    avatar_url VARCHAR(500),
    
    -- Authentication provider info
    provider VARCHAR(50) NOT NULL, -- github, azuread, oidc
    provider_user_id VARCHAR(255) NOT NULL,
    provider_data JSONB NOT NULL DEFAULT '{}',
    
    -- User preferences and settings
    preferences JSONB NOT NULL DEFAULT '{}',
    timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Account status
    is_active BOOLEAN NOT NULL DEFAULT true,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    last_login_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(provider, provider_user_id)
);

CREATE INDEX idx_users_email ON user_management.users(email);
CREATE INDEX idx_users_provider ON user_management.users(provider, provider_user_id);
CREATE INDEX idx_users_active ON user_management.users(is_active);

-- Tenant memberships - user access to tenants
CREATE TABLE user_management.tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES user_management.users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES control_plane.tenants(id) ON DELETE CASCADE,
    role VARCHAR(100) NOT NULL DEFAULT 'developer',
    permissions JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT true,
    invited_by UUID REFERENCES user_management.users(id),
    invited_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    joined_at TIMESTAMP WITH TIME ZONE,
    
    UNIQUE(user_id, tenant_id),
    CONSTRAINT valid_role CHECK (role IN ('owner', 'admin', 'developer', 'viewer'))
);

CREATE INDEX idx_tenant_memberships_user ON user_management.tenant_memberships(user_id);
CREATE INDEX idx_tenant_memberships_tenant ON user_management.tenant_memberships(tenant_id);
CREATE INDEX idx_tenant_memberships_active ON user_management.tenant_memberships(is_active);

-- Audit System Schema  
-- ===================

-- Comprehensive audit log for all platform operations
CREATE TABLE audit_system.audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES control_plane.tenants(id) ON DELETE SET NULL,
    user_id UUID REFERENCES user_management.users(id) ON DELETE SET NULL,
    
    -- Event identification
    event_type VARCHAR(100) NOT NULL, -- create, update, delete, login, etc.
    resource_type VARCHAR(100), -- Application, Team, Repository, etc.
    resource_id UUID,
    resource_name VARCHAR(255),
    
    -- Event details
    action VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Request context
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),
    session_id VARCHAR(100),
    
    -- Data changes (for update operations)
    old_values JSONB,
    new_values JSONB,
    
    -- Outcome and metadata
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    duration_ms INTEGER,
    additional_data JSONB DEFAULT '{}',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for audit queries (partition-friendly)
CREATE INDEX idx_audit_tenant_time ON audit_system.audit_log(tenant_id, created_at);
CREATE INDEX idx_audit_user_time ON audit_system.audit_log(user_id, created_at);
CREATE INDEX idx_audit_event_type ON audit_system.audit_log(event_type, created_at);
CREATE INDEX idx_audit_resource ON audit_system.audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_request ON audit_system.audit_log(request_id);

-- Git Integration Schema
-- =====================

-- Repository connections and webhook management
CREATE TABLE git_integration.repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES control_plane.tenants(id) ON DELETE CASCADE,
    application_id UUID REFERENCES resource_management.applications(id) ON DELETE CASCADE,
    
    -- Repository identification
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(500) NOT NULL, -- org/repo format
    description TEXT,
    
    -- Provider information
    provider VARCHAR(50) NOT NULL DEFAULT 'github',
    provider_repo_id VARCHAR(100) NOT NULL,
    
    -- Repository URLs and access
    clone_url VARCHAR(500) NOT NULL,
    web_url VARCHAR(500) NOT NULL,
    default_branch VARCHAR(100) NOT NULL DEFAULT 'main',
    
    -- Repository settings
    visibility VARCHAR(20) NOT NULL DEFAULT 'private',
    archived BOOLEAN NOT NULL DEFAULT false,
    
    -- Access control
    access_config JSONB NOT NULL DEFAULT '{}',
    branch_protection JSONB NOT NULL DEFAULT '{}',
    
    -- Webhook configuration
    webhook_id VARCHAR(100),
    webhook_secret_hash VARCHAR(255),
    webhook_events JSONB NOT NULL DEFAULT '[]',
    
    -- Status and metadata
    last_push_at TIMESTAMP WITH TIME ZONE,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    sync_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    
    CONSTRAINT valid_provider CHECK (provider IN ('github', 'gitlab', 'bitbucket')),
    CONSTRAINT valid_visibility CHECK (visibility IN ('public', 'private', 'internal')),
    CONSTRAINT valid_sync_status CHECK (sync_status IN ('pending', 'synced', 'error')),
    UNIQUE(tenant_id, provider, provider_repo_id)
);

CREATE INDEX idx_repositories_tenant ON git_integration.repositories(tenant_id);
CREATE INDEX idx_repositories_application ON git_integration.repositories(application_id);
CREATE INDEX idx_repositories_provider ON git_integration.repositories(provider, provider_repo_id);
CREATE INDEX idx_repositories_sync ON git_integration.repositories(sync_status);

-- Webhook events and processing
CREATE TABLE git_integration.webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repository_id UUID NOT NULL REFERENCES git_integration.repositories(id) ON DELETE CASCADE,
    
    -- Event identification
    event_type VARCHAR(100) NOT NULL, -- push, pull_request, issues, etc.
    event_id VARCHAR(100) NOT NULL, -- External event ID from provider
    
    -- Event payload
    payload JSONB NOT NULL DEFAULT '{}',
    headers JSONB NOT NULL DEFAULT '{}',
    
    -- Processing status
    processed BOOLEAN NOT NULL DEFAULT false,
    processing_attempts INTEGER NOT NULL DEFAULT 0,
    processing_error TEXT,
    processed_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(repository_id, event_id)
);

CREATE INDEX idx_webhook_events_repository ON git_integration.webhook_events(repository_id);
CREATE INDEX idx_webhook_events_unprocessed ON git_integration.webhook_events(processed, created_at) WHERE NOT processed;
CREATE INDEX idx_webhook_events_type ON git_integration.webhook_events(event_type, created_at);

-- Update timestamp triggers
-- ========================

-- Function to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers to all tables with updated_at columns
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON control_plane.tenants 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_applications_updated_at BEFORE UPDATE ON resource_management.applications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON resource_management.teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    
CREATE TRIGGER update_resources_updated_at BEFORE UPDATE ON resource_management.resources
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON user_management.users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_repositories_updated_at BEFORE UPDATE ON git_integration.repositories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Initial data
-- ============

-- Create default system tenant for platform operations
INSERT INTO control_plane.tenants (
    id,
    name, 
    display_name, 
    description,
    database_name,
    status,
    settings
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'system',
    'System Tenant', 
    'Internal platform operations and administration',
    'platform_system',
    'active',
    '{"system_tenant": true, "auto_created": true}'
) ON CONFLICT (name) DO NOTHING;

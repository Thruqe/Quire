-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Workspaces
CREATE TABLE IF NOT EXISTS workspaces (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  owner_id UUID REFERENCES users(id),
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Pages
CREATE TABLE IF NOT EXISTS pages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workspace_id UUID REFERENCES workspaces(id),
  parent_id UUID REFERENCES pages(id),
  title TEXT NOT NULL DEFAULT '',
  icon TEXT,
  cover TEXT,
  properties JSONB DEFAULT '{}',
  created_by UUID REFERENCES users(id),
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Yjs update log
CREATE TABLE IF NOT EXISTS doc_updates (
  id BIGSERIAL PRIMARY KEY,
  page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
  update_data BYTEA NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Yjs snapshots
CREATE TABLE IF NOT EXISTS doc_snapshots (
  page_id UUID PRIMARY KEY REFERENCES pages(id) ON DELETE CASCADE,
  snapshot BYTEA NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Database schemas
CREATE TABLE IF NOT EXISTS db_schemas (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
  fields JSONB NOT NULL DEFAULT '[]'
);

-- Permissions
CREATE TABLE IF NOT EXISTS page_permissions (
  page_id UUID REFERENCES pages(id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  role TEXT CHECK (role IN ('viewer', 'editor', 'admin')),
  PRIMARY KEY (page_id, user_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_pages_parent ON pages(parent_id);
CREATE INDEX IF NOT EXISTS idx_pages_workspace ON pages(workspace_id);
CREATE INDEX IF NOT EXISTS idx_doc_updates_page ON doc_updates(page_id);
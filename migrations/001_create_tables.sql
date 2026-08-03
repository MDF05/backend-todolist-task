-- Migration: 001_create_tables.sql
-- Creates users and tasks tables with proper indexes

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Users table ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255)             NOT NULL,
    email      VARCHAR(255) UNIQUE      NOT NULL,
    password   VARCHAR(255)             NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_users_email     ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- ─── Tasks table ──────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255)                    NOT NULL,
    description TEXT,
    status      VARCHAR(20)                     NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'completed')),
    due_date    DATE,
    created_at  TIMESTAMP WITH TIME ZONE        DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE        DEFAULT NOW(),
    deleted_at  TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_tasks_status     ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);

-- Full-text search index on title and description
CREATE INDEX IF NOT EXISTS idx_tasks_search
    ON tasks USING GIN(to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(description, '')));

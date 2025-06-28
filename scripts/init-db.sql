-- Aura CLI Database Schema
-- This file initializes the SQLite database for Aura CLI

PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 268435456; -- 256MB

-- Create bookmarks table
CREATE TABLE IF NOT EXISTS bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    alias TEXT UNIQUE NOT NULL,
    path TEXT NOT NULL,
    description TEXT,
    tags TEXT, -- JSON array of tags
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    access_count INTEGER DEFAULT 0
);

-- Create navigation history table
CREATE TABLE IF NOT EXISTS navigation_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create user preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create command history table for AI context
CREATE TABLE IF NOT EXISTS command_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    working_directory TEXT NOT NULL,
    exit_code INTEGER,
    execution_time_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create project context cache table
CREATE TABLE IF NOT EXISTS project_contexts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT UNIQUE NOT NULL,
    context_type TEXT NOT NULL, -- git, node, python, etc.
    metadata TEXT, -- JSON metadata
    last_detected DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_bookmarks_alias ON bookmarks(alias);
CREATE INDEX IF NOT EXISTS idx_bookmarks_path ON bookmarks(path);
CREATE INDEX IF NOT EXISTS idx_bookmarks_tags ON bookmarks(tags);
CREATE INDEX IF NOT EXISTS idx_navigation_history_path ON navigation_history(path);
CREATE INDEX IF NOT EXISTS idx_navigation_history_accessed_at ON navigation_history(accessed_at);
CREATE INDEX IF NOT EXISTS idx_user_preferences_key ON user_preferences(key);
CREATE INDEX IF NOT EXISTS idx_command_history_working_directory ON command_history(working_directory);
CREATE INDEX IF NOT EXISTS idx_command_history_created_at ON command_history(created_at);
CREATE INDEX IF NOT EXISTS idx_project_contexts_path ON project_contexts(path);
CREATE INDEX IF NOT EXISTS idx_project_contexts_context_type ON project_contexts(context_type);

-- Insert default preferences
INSERT OR IGNORE INTO user_preferences (key, value) VALUES 
    ('ai_model', 'gpt-4.1-nano'),
    ('auto_bookmark_threshold', '5'),
    ('history_retention_days', '90'),
    ('default_editor', 'vim'),
    ('fuzzy_search_enabled', 'true');

-- Create triggers for updated_at timestamps
CREATE TRIGGER IF NOT EXISTS bookmarks_updated_at 
    AFTER UPDATE ON bookmarks
BEGIN
    UPDATE bookmarks SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS user_preferences_updated_at 
    AFTER UPDATE ON user_preferences
BEGIN
    UPDATE user_preferences SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create trigger to update bookmark access count
CREATE TRIGGER IF NOT EXISTS bookmark_access_counter
    AFTER INSERT ON navigation_history
    WHEN NEW.success = TRUE
BEGIN
    UPDATE bookmarks 
    SET access_count = access_count + 1,
        updated_at = CURRENT_TIMESTAMP
    WHERE path = NEW.path;
END;

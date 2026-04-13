-- Add a dependencies column to the plugins table so each installed plugin can
-- record the plugin IDs it declared as required at install time.
ALTER TABLE plugins
    ADD COLUMN IF NOT EXISTS dependencies JSONB NOT NULL DEFAULT '[]';

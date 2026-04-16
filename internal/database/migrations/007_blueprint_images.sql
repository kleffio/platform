ALTER TABLE blueprints ADD COLUMN images JSONB NOT NULL DEFAULT '{}'::jsonb;

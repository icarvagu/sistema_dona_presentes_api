ALTER TABLE purchase_history
    ADD COLUMN IF NOT EXISTS user_name TEXT NOT NULL DEFAULT 'Sistema';

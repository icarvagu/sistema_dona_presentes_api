ALTER TABLE sale_items
  ADD COLUMN IF NOT EXISTS engravings JSONB DEFAULT '[]'::jsonb;


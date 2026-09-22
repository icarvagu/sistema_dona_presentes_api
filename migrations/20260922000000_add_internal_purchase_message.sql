ALTER TABLE purchase_requests
    ADD COLUMN IF NOT EXISTS buyer_message TEXT NOT NULL DEFAULT '';

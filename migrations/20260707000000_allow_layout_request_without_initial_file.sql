-- +goose Up
ALTER TABLE layout_requests DROP CONSTRAINT IF EXISTS layout_requests_file_mode_check;
ALTER TABLE layout_requests ADD CONSTRAINT layout_requests_file_mode_check
    CHECK (file_mode IN ('none','common','separate'));

-- +goose Down
UPDATE layout_requests SET file_mode='common' WHERE file_mode='none';
ALTER TABLE layout_requests DROP CONSTRAINT IF EXISTS layout_requests_file_mode_check;
ALTER TABLE layout_requests ADD CONSTRAINT layout_requests_file_mode_check
    CHECK (file_mode IN ('common','separate'));

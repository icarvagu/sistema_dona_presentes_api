-- +goose Up
ALTER TABLE sales ADD COLUMN layout_urls TEXT[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE sales DROP COLUMN layout_urls;

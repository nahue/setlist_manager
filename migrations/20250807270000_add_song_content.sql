-- +goose Up
ALTER TABLE songs ADD COLUMN content TEXT;

-- +goose Down
ALTER TABLE songs DROP COLUMN content;

-- +goose Up
UPDATE orders SET user_uuid = '550e8400-e29b-41d4-a716-446655440000' WHERE user_uuid IS NULL;

-- +goose Down
UPDATE orders SET user_uuid = NULL WHERE user_uuid = '550e8400-e29b-41d4-a716-446655440000';


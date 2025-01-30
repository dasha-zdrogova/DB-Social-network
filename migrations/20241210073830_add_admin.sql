-- +goose Up
-- +goose StatementBegin
INSERT INTO users (username, password_hash, role)
VALUES (
        'admin',
        crypt('admin123', gen_salt('bf')),
        'admin'
    ) ON CONFLICT (username) DO NOTHING;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DELETE FROM users
WHERE username = 'admin';
-- +goose StatementEnd
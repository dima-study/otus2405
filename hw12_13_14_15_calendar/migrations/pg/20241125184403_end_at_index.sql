-- +goose Up
-- +goose StatementBegin
CREATE INDEX "end_at" on events (upper(time));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX "end_at";
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE INDEX "need_notify" on events ((lower(time) - notify_before * '1 day'::interval)) WHERE notify_before>0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX "need_notify";
-- +goose StatementEnd

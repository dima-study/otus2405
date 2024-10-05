-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION btree_gist;

CREATE TABLE "events" (
  "id"            bigserial       NOT NULL PRIMARY KEY,
  "event_id"      uuid            NOT NULL,
  "owner_id"      uuid            NOT NULL,
  "time"          tsrange         NOT NULL,
  "title"         varchar(128)    NOT NULL,
  "description"   text                NULL,
  "notify_before" int             NOT NULL DEFAULT 0,

  CONSTRAINT "uniq_owner_event_id" UNIQUE ("owner_id", "event_id"),
  CONSTRAINT "positive_notify_before" CHECK ("notify_before" >= 0),
  CONSTRAINT "no_time_overlap" EXCLUDE USING GIST ("owner_id" WITH =, "time" WITH &&)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "events";

DROP EXTENSION btree_gist;
-- +goose StatementEnd

-- +migrate Up
ALTER TABLE "public"."task" ADD COLUMN "position" int4 NOT NULL DEFAULT 0;

-- +migrate Down
ALTER TABLE "public"."task" DROP COLUMN "position";


-- +migrate Up
CREATE TABLE "public"."task"
(
	"id"           uuid                                NOT NULL DEFAULT uuid_generate_v4(),
	"list_id"      uuid                                NOT NULL,
	"description"  text COLLATE "pg_catalog"."default" NOT NULL,
	"is_completed" bool                                NOT NULL DEFAULT false,
	"created_at"   timestamptz(6)                      NOT NULL DEFAULT now(),
	"updated_at"   timestamptz(6)                      NOT NULL DEFAULT now(),
	"completed_at" timestamptz(6)
)
;
ALTER TABLE "public"."task" OWNER TO "postgres";
ALTER TABLE "public"."task" ADD CONSTRAINT "task_pkey" PRIMARY KEY ("id");
ALTER TABLE "public"."task" ADD CONSTRAINT "task_list_id_fkey" FOREIGN KEY ("list_id") REFERENCES "public"."list" ("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- +migrate Down
DROP TABLE IF EXISTS "public"."task";


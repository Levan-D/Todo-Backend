-- +migrate Up
CREATE TABLE "public"."user"
(
	"id"                    uuid                                        NOT NULL DEFAULT uuid_generate_v4(),
	"avatar"                text COLLATE "pg_catalog"."default",
	"email"                 varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
	"password"              varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
	"first_name"            varchar(120) COLLATE "pg_catalog"."default" NOT NULL,
	"last_name"             varchar(160) COLLATE "pg_catalog"."default" NOT NULL,
	"is_verified"           bool                                                 DEFAULT false,
	"reset_password_token"  varchar(255) COLLATE "pg_catalog"."default",
	"reset_password_expire" timestamptz(6),
	"created_at"            timestamptz(6)                              NOT NULL DEFAULT now(),
	"updated_at"            timestamptz(6)                              NOT NULL DEFAULT now(),
	"verified_at"           timestamptz(6)
)
;
ALTER TABLE "public"."user" OWNER TO "postgres";
ALTER TABLE "public"."user" ADD CONSTRAINT "user_email_key" UNIQUE ("email");
ALTER TABLE "public"."user" ADD CONSTRAINT "user_pkey" PRIMARY KEY ("id");

-- +migrate Down
DROP TABLE IF EXISTS "public"."user";


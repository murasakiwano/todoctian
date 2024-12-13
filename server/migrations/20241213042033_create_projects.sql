CREATE TABLE "projects" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamp NOT NULL DEFAULT now(),
  "name" text NOT NULL UNIQUE,
  PRIMARY KEY ("id")
);

-- Create index "project_name" to table: "projects"
CREATE INDEX "project_name" ON "public"."projects" ("name");

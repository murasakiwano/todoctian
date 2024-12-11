-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";

-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'standard public schema';

-- Enable fuzzystrmatch extension for fuzzy searching
CREATE EXTENSION IF NOT EXISTS fuzzystrmatch;

-- Create "projects" table
CREATE TABLE "public"."projects" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamp NOT NULL DEFAULT now(),
  "name" text NOT NULL UNIQUE,
  PRIMARY KEY ("id")
);

-- Create index "project_name" to table: "projects"
CREATE INDEX "project_name" ON "public"."projects" ("name");

-- Create "tasks" table
CREATE TABLE "public"."tasks" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamp NOT NULL DEFAULT now(),
  "parent_task_id" uuid NULL,
  "project_id" uuid NOT NULL,
  "status" text NOT NULL DEFAULT 'Pending',
  "order" integer NOT NULL DEFAULT 0,
  "name" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "tasks_project_id_parent_task_id_order_key" UNIQUE ("project_id", "parent_task_id", "order"),
  CONSTRAINT "tasks_parent_task_id_fkey" FOREIGN KEY ("parent_task_id") REFERENCES "public"."tasks" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "tasks_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "tasks_order_check" CHECK ("order" >= 0),
  CONSTRAINT "tasks_status_check" CHECK (status = ANY (ARRAY['pending'::text, 'completed'::text]))
);

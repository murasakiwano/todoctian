env "local" {
  src = "file://migrations/schema.sql"
  url = "postgres://postgres:pass@db:5432/todoctian?sslmode=disable"
  dev = "docker://postgres/16/dev?search_path=public"
}

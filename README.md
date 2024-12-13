# _Todoctian_

This project is a simple todo-list API developed in Go as part of a technical
challenge for [TRACTIAN](https://tractian.com).

## Features

- Projects
  - A project is a collection of todo items
  - You can add, rename, and delete projects
  - Deleting a project deletes all its tasks
- Todos (tasks)
  - Each task must belong to a project
  - A task may or may not have subtasks
    - When you complete a task, all its subtasks are completed also
    - When a task is marked as pending, its parent task is also marked as pending
  - All tasks in the same level (e.g., at the root of a project) have a specific order
    - You can re-order these tasks as you please

## API Documentation

The OpenAPI spec is located at [openapi.yaml](./server/api/openapi.yaml). It
defines relevant endpoints, schemas, models, request and response formats.
To view the documentation in a more readable format, you can use something like
[Swagger UI](https://swagger.io/tools/swagger-ui/) or [Redoc](https://github.com/Redocly/redoc).
For instructions on each tool, see their respective URLs.

## Prerequisites

- [Go](https://golang.org/dl/) 1.23 or higher
- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Running locally

Clone the repository:

```shell
git clone https://github.com/murasakiwano/todoctian
```

There is a `compose.yaml` file in the root of the repository containing the
definition of a server and a db service. The server connects to the Postgres
database in order to run the API. In between those two services, a migration
service spins up. It sets up the database schema for the server. Simply run:

```shell
docker compose up --build server
```

Then everything should start running.

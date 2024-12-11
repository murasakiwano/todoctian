-- name: CreateProject :exec
INSERT INTO projects (
  id, name, created_at
) VALUES (
  $1, $2, $3
);

-- name: GetProject :one
SELECT * FROM projects
WHERE id = $1 LIMIT 1;

-- name: GetProjectByName :one
SELECT * FROM projects
WHERE name = $1 LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects
ORDER BY name;

-- name: RenameProject :exec
UPDATE projects
SET name = $2
WHERE id = $1;

-- name: DeleteProject :one
DELETE FROM projects
WHERE id = $1
RETURNING *;

-- name: CreateTask :exec
INSERT INTO tasks (
  id, project_id, name, status, "order", parent_task_id, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
);

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1 LIMIT 1;

-- name: GetSubtasksDirect :many
SELECT * FROM tasks
WHERE parent_task_id = $1;

-- name: GetSubtasksDeep :many
WITH RECURSIVE subtasks AS (
  -- Base case: Direct children of the specified parent task
  SELECT * FROM tasks ts
  WHERE ts.parent_task_id = $1

  UNION

  -- Recursive step: For each found subtask, find its own children
  SELECT t.* FROM tasks t
  INNER JOIN subtasks st ON t.parent_task_id = st.id
)
SELECT * FROM subtasks;

-- name: GetTasksByProject :many
SELECT * FROM tasks
WHERE project_id = $1;

-- name: GetTasksInProjectRoot :many
SELECT * FROM tasks
WHERE project_id = $1 AND parent_task_id IS NULL;

-- name: GetTasksByStatus :many
SELECT * FROM tasks
WHERE project_id = $1 AND status = $2;

-- name: RenameTask :exec
UPDATE tasks
SET name = $2
WHERE id = $1;

-- name: UpdateTaskOrder :exec
UPDATE tasks
SET "order" = $2
WHERE id = $1;

-- name: UpdateTaskStatus :exec
UPDATE tasks
SET "status" = $2
WHERE id = $1;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- WARN: the following two queries should be used together
-- in the scope of a transaction!

-- name: OffsetTaskOrders :exec
UPDATE tasks
SET "order" = "order" + 1000
WHERE project_id = @project_id::uuid
  AND (
    (parent_task_id IS NULL AND @parent_task_id::uuid IS NULL) OR
    (parent_task_id = @parent_task_id::uuid)
  );

-- name: BatchUpdateTaskOrders :batchexec
UPDATE tasks
SET "order" = $2
WHERE tasks.id = $1;

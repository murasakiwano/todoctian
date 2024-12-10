package project

import (
	"github.com/google/uuid"
	"github.com/murasakiwano/todoctian/server/db"
	"github.com/murasakiwano/todoctian/server/internal"
)

func AdaptDBProjectToProjectModel(projectDB db.Project) (Project, error) {
	projectID := internal.EncodeUUID(projectDB.ID.Bytes)
	id, err := uuid.Parse(projectID)
	if err != nil {
		return Project{}, err
	}

	createdAt := projectDB.CreatedAt.Time

	return Project{
		ID:        id,
		CreatedAt: createdAt,
		Name:      projectDB.Name,
	}, nil
}

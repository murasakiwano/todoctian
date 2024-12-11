package project

import (
	"github.com/murasakiwano/todoctian/server/db"
	"github.com/murasakiwano/todoctian/server/internal"
)

func ProjectDBToProjectModel(projectDB db.Project) (Project, error) {
	projectID, err := internal.EncodeUUID(projectDB.ID.Bytes)
	if err != nil {
		return Project{}, err
	}

	createdAt := projectDB.CreatedAt.Time

	return Project{
		ID:        projectID,
		CreatedAt: createdAt,
		Name:      projectDB.Name,
	}, nil
}

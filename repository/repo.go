package repository

import (
	"context"
)

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ForksCount  int    `json:"forksCount"`
}

type Repository interface {
	Fetch(context.Context, int) ([]Project, error)
}

type inMemoryRepository struct {
	fakeProjects []Project
}

func NewInMemoryRepository() *inMemoryRepository {
	ir := inMemoryRepository{}
	ir.fakeProjects = []Project{
		{Name: "Boner project", ForksCount: 0},
		{Name: "grup", ForksCount: 0},
		{Name: "easy", ForksCount: 2},
		{Name: "slothbeast", ForksCount: 4},
		{Name: "sspssptest", ForksCount: 0},
		{Name: "hcs_utils", ForksCount: 1},
		{Name: "K", ForksCount: 1},
		{Name: "Heroes of Wesnoth", ForksCount: 5},
		{Name: "Leiningen", ForksCount: 1},
		{Name: "TearDownWalls", ForksCount: 5},
	}

	return &ir
}

func (ir *inMemoryRepository) Fetch(ctx context.Context, n int) ([]Project, error) {
	repoChan := make(chan []Project)

	go func() {
		repoChan <- ir.fakeProjects
	}()

	var err error
	latestProjects := []Project{}

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case latestProjects = <-repoChan:
	}
	return latestProjects, err
}

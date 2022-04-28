package repository

import (
	"context"
	"fmt"

	"github.com/machinebox/graphql"
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
		if n > len(ir.fakeProjects) {
			n = len(ir.fakeProjects)
		}

		repoChan <- ir.fakeProjects[:n]
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

type gitlabRepository struct {
	apiEndpoint string
}

func NewGitlabRepository(endPoint string) *gitlabRepository {
	return &gitlabRepository{apiEndpoint: endPoint}
}

func (gr *gitlabRepository) Fetch(ctx context.Context, n int) ([]Project, error) {
	type projects struct {
		Nodes []Project `json:"nodes"`
	}

	type response struct {
		Projects projects `json:"projects"`
	}

	query := `
	query last_projects($n: Int = 5) {
	projects(last: $n) {
	nodes {
	name
	description
	forksCount
	}
	}
	}
	`
	request := graphql.NewRequest(query)
	request.Var("n", n)

	client := graphql.NewClient(gr.apiEndpoint)

	var resp response
	if err := client.Run(ctx, request, &resp); err != nil {
		return nil, fmt.Errorf("graphql client error: %w", err)
	}

	var latestProjects []Project
	for _, proj := range resp.Projects.Nodes {
		latestProjects = append(latestProjects, proj)
	}

	return latestProjects, nil
}

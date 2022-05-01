package repository

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/machinebox/graphql"
	"golang.org/x/oauth2"
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

func NewInMemoryRepository(projects ...Project) *inMemoryRepository {
	if len(projects) == 0 {
		projects = []Project{
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
	}

	return &inMemoryRepository{projects}
}

func (ir *inMemoryRepository) Fetch(ctx context.Context, count int) ([]Project, error) {
	repoChan := make(chan []Project)

	go func() {
		if count > len(ir.fakeProjects) {
			count = len(ir.fakeProjects)
		}

		repoChan <- ir.fakeProjects[:count]
	}()

	var err error

	latestProjects := []Project{}

	select {
	case <-ctx.Done():
		err = fmt.Errorf("repository fetch aborted: %w", ctx.Err())
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

func (gr *gitlabRepository) Fetch(ctx context.Context, count int) ([]Project, error) {
	type projects struct {
		Nodes []Project `json:"nodes"`
	}

	type response struct {
		Projects projects `json:"projects"`
	}

	query := `
	query last_projects($n: Int) {
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
	request.Var("n", count)

	httpClient := http.DefaultClient

	if graphqlToken := os.Getenv("FORKSCOUNT_GRAPHQL_TOKEN"); graphqlToken != "" {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: graphqlToken},
		)

		httpClient = oauth2.NewClient(ctx, src)
	}

	client := graphql.NewClient(gr.apiEndpoint, graphql.WithHTTPClient(httpClient))

	var resp response
	if err := client.Run(ctx, request, &resp); err != nil {
		return nil, fmt.Errorf("repository fetch error: %w", err)
	}

	var latestProjects []Project
	for _, proj := range resp.Projects.Nodes {
		latestProjects = append(latestProjects, proj)
	}

	return latestProjects, nil
}

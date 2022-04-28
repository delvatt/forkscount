package service

import (
	"context"
	"log"
	"strings"

	"github.com/delvatt/forkscount/repository"
)

type ApiResponse struct {
	Names    string `json:"projectNames"`
	ForksSum int    `json:"forksSum"`
}

func GetLatestProject(repo repository.Repository, n int) ApiResponse {
	ctx := context.Background()
	projects, err := repo.Fetch(ctx, n)
	if err != nil {
		log.Fatal(err)
	}

	var sum int
	var names []string
	for _, project := range projects {
		names = append(names, project.Name)
		sum += project.ForksCount
	}

	return ApiResponse{
		Names:    strings.Join(names, ","),
		ForksSum: sum}
}

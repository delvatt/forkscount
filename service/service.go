package service

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/delvatt/forkscount/repository"
)

const defaultCount = 5
const maxCount = 100

type ApiResponse struct {
	Names    string `json:"names"`
	ForksSum int    `json:"forksSum"`
}

func GetLatestProject(repo repository.Repository, lastCount int) ApiResponse {
	ctx := context.Background()
	projects, err := repo.Fetch(ctx, lastCount)
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

func GetLatestProjectJSONHandler(w http.ResponseWriter, r *http.Request) {
	repo, lastCount := preProcessRequest(r)
	jsonData, err := json.Marshal(GetLatestProject(repo, lastCount))
	if err != nil {
		log.Fatal(err)
	}

	w.Write(jsonData)
}

func preProcessRequest(r *http.Request) (repository.Repository, int) {
	n := r.URL.Query().Get("n")

	lastCount, err := strconv.Atoi(n)
	if err != nil || lastCount < 0 {
		// log that we are falling back on the default value
		lastCount = defaultCount
	}

	if lastCount > maxCount {
		lastCount = maxCount
	}

	repo := repository.NewInMemoryRepository()

	return repo, lastCount
}

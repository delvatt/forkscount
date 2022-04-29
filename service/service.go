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

func GetLatestProject(ctx context.Context, repo repository.Repository, lastCount int) ApiResponse {
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

func GetLatestProjectJSONHandler(repo repository.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lastCount := preProcessRequest(r)
		jsonData, err := json.Marshal(GetLatestProject(r.Context(), repo, lastCount))
		if err != nil {
			log.Fatal(err)
		}

		w.Write(jsonData)
	}
}

func preProcessRequest(r *http.Request) int {
	n := r.URL.Query().Get("n")

	lastCount, err := strconv.Atoi(n)
	if err != nil || lastCount < 0 {
		// log that we are falling back on the default value
		lastCount = defaultCount
	}

	if lastCount > maxCount {
		lastCount = maxCount
	}

	return lastCount
}

func NewService(addr string) *http.Server {
	http.HandleFunc("/", GetLatestProjectJSONHandler(repository.NewGitlabRepository("https://gitlab.com/api/graphql")))

	return &http.Server{
		Addr:    addr,
		Handler: nil,
	}
}

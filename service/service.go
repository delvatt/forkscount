package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/delvatt/forkscount/repository"
)

const defaultCount = 5
const maxCount = 100

const gitlabGraphqlEndpoint = "https://gitlab.com/api/graphql"

type ApiResponse struct {
	Names    string `json:"names"`
	ForksSum int    `json:"forksSum"`
}

func GetLatestProject(ctx context.Context, repo repository.Repository, lastCount int) (ApiResponse, error) {
	projects, err := repo.Fetch(ctx, lastCount)
	if err != nil {
		return ApiResponse{}, err
	}

	var sum int
	var names []string
	for _, project := range projects {
		names = append(names, project.Name)
		sum += project.ForksCount
	}

	return ApiResponse{
		Names:    strings.Join(names, ","),
		ForksSum: sum}, nil
}

func GetLatestProjectJSONHandler(repo repository.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lastCount := getLastCount(r)

		data, err := GetLatestProject(r.Context(), repo, lastCount)
		if err != nil {
			log.Printf("failed to fetch data from repository: %v", err)
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				http.Error(w, err.Error(), http.StatusRequestTimeout)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("failed to marshal repository data into json: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(jsonData)
	}
}

func getLastCount(r *http.Request) int {
	var lastCount = defaultCount
	n := r.URL.Query().Get("n")

	if n != "" {
		val, err := strconv.Atoi(n)
		if err != nil || val < 0 {
			log.Printf("invalid lastCount parameter %q, falling back to default value %d\n", n, defaultCount)
		} else {
			lastCount = val
		}
	}

	if lastCount > maxCount {
		lastCount = maxCount
	}

	return lastCount
}

func NewService(addr string) *http.Server {
	http.HandleFunc("/", GetLatestProjectJSONHandler(repository.NewGitlabRepository(os.Getenv("FORKSCOUNT_GRAPHQL_SERVER_ADDR"))))

	return &http.Server{
		Addr:    addr,
		Handler: nil,
	}
}

func init() {
	if graphqlAddr := os.Getenv("FORKSCOUNT_GRAPHQL_SERVER_ADDR"); graphqlAddr == "" {
		log.Printf("missing %q env var, defaulting to %q\n", "FORKSCOUNT_GRAPHQL_SERVER_ADDR", gitlabGraphqlEndpoint)
		log.Printf("please consider setting %q env var\n", "FORKSCOUNT_GRAPHQL_SERVER_ADDR")
		os.Setenv("FORKSCOUNT_GRAPHQL_SERVER_ADDR", gitlabGraphqlEndpoint)
	}
}

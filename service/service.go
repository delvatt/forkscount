package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

var serviceLogger = log.New(os.Stdout, "[SERVICE]:", log.LstdFlags)

type APIResponse struct {
	Names    string `json:"names"`
	ForksSum int    `json:"forksSum"`
}

func GetLatestProject(ctx context.Context, repo repository.Repository, lastCount int) (APIResponse, error) {
	projects, err := repo.Fetch(ctx, lastCount)
	if err != nil {
		return APIResponse{}, fmt.Errorf("fail to fetch data from repository: %w", err)
	}

	var sum int

	var names []string

	for _, project := range projects {
		names = append(names, project.Name)
		sum += project.ForksCount
	}

	return APIResponse{
		Names:    strings.Join(names, ","),
		ForksSum: sum}, nil
}

func GetLatestProjectJSONHandler(repo repository.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		lastCount := getLastCount(r)

		data, err := GetLatestProject(r.Context(), repo, lastCount)
		if err != nil {
			logMsg("failed to fetch latest project(s): %v", err)

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
			logMsg("failed to marshal repository data into json: %v", err)
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
			logMsg("invalid lastCount parameter %q, falling back to default value %d\n", n, defaultCount)
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
	if os.Getenv("FORKSCOUNT_GRAPHQL_TOKEN") == "" {
		logMsg("missing %q env var for authenticated graphql queries\n",
			"FORKSCOUNT_GRAPHQL_TOKEN")
		logMsg("you can configure your authentication token for the remote graphql repository by setting %q env var\n",
			"FORKSCOUNT_GRAPHQL_TOKEN")
	}

	var graphqlAddr string

	if graphqlAddr = os.Getenv("FORKSCOUNT_GRAPHQL_SERVER_ADDR"); graphqlAddr == "" {
		logMsg("missing %q env var, defaulting to %q\n",
			"FORKSCOUNT_GRAPHQL_SERVER_ADDR", gitlabGraphqlEndpoint)
		logMsg("please consider setting %q env var\n", "FORKSCOUNT_GRAPHQL_SERVER_ADDR")

		graphqlAddr = gitlabGraphqlEndpoint
	}

	http.HandleFunc("/", GetLatestProjectJSONHandler(repository.NewGitlabRepository(graphqlAddr)))

	return &http.Server{
		Addr:    addr,
		Handler: nil,
	}
}

func logMsg(format string, v ...any) {
	logSuffix := os.Getenv("FORKSCOUNT_LOGGING_ENABLED")

	if logSuffix != "" {
		logTarget, err := os.OpenFile(logSuffix, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			serviceLogger.Printf("fail to create service log file: %v", err)

			return
		}
		defer logTarget.Close()

		w := serviceLogger.Writer()

		serviceLogger.SetOutput(logTarget)

		defer serviceLogger.SetOutput(w)
	}

	serviceLogger.Printf(format, v...)
}

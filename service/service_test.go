package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delvatt/forkscount/repository"
	"github.com/delvatt/forkscount/service"
)

func TestServiceCore(t *testing.T) {
	t.Parallel()

	var fakeProjects = []repository.Project{
		{Name: "hcs_utils", ForksCount: 1},
		{Name: "K", ForksCount: 1},
		{Name: "Heroes of Wesnoth", ForksCount: 5},
		{Name: "Leiningen", ForksCount: 1},
		{Name: "TearDownWalls", ForksCount: 5},
	}

	want := service.APIResponse{
		Names:    "hcs_utils,K,Heroes of Wesnoth,Leiningen,TearDownWalls",
		ForksSum: 13,
	}
	got, err := service.GetLatestProject(context.Background(), repository.NewInMemoryRepository(fakeProjects...), 5)

	if err != nil {
		t.Error("expected no error, but got one")
	}

	if want != got {
		t.Errorf("expected %v, but got %v", want, got)
	}
}

func TestServiceHandler(t *testing.T) {
	t.Parallel()

	var server http.HandlerFunc
	server = service.GetLatestProjectJSONHandler(repository.NewInMemoryRepository())

	tests := []struct {
		name        string
		fakeRequest *http.Request
		expected    string
	}{
		{
			name:        "WithNoLimit",
			fakeRequest: httptest.NewRequest(http.MethodGet, "/", nil),
			expected:    `{"names":"Boner project,grup,easy,slothbeast,sspssptest","forksSum":6}`,
		},
		{
			name: "WithLimit",
			fakeRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				q := req.URL.Query()
				q.Add("n", "10")
				req.URL.RawQuery = q.Encode()

				return req
			}(),
			expected: `{"names":"Boner project,grup,easy,slothbeast,sspssptest,hcs_utils,K,Heroes of Wesnoth,Leiningen,TearDownWalls","forksSum":19}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fakeResponse := httptest.NewRecorder()
			server.ServeHTTP(fakeResponse, test.fakeRequest)

			jsonString := fakeResponse.Body.String()
			if jsonString != test.expected {
				t.Errorf("expected %s, but got %s", test.expected, jsonString)
			}
		})
	}
}

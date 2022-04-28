package service_test

import (
	"testing"

	"github.com/delvatt/forkscount/repository"
	"github.com/delvatt/forkscount/service"
)

var fakeProjects = []repository.Project{
	{Name: "hcs_utils", ForksCount: 1},
	{Name: "K", ForksCount: 1},
	{Name: "Heroes of Wesnoth", ForksCount: 5},
	{Name: "Leiningen", ForksCount: 1},
	{Name: "TearDownWalls", ForksCount: 5},
}

func TestServiceCore(t *testing.T) {
	want := service.ApiResponse{
		Names:    "hcs_utils,K,Heroes of Wesnoth,Leiningen,TearDownWalls",
		ForksSum: 13,
	}
	got := service.GetLatestProject(repository.NewInMemoryRepository(fakeProjects), 5)

	if want != got {
		t.Errorf("expected %v, but got %v", want, got)
	}
}

package repository_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/delvatt/forkscount/repository"
)

var fakeProjects = []repository.Project{
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

func TestFetch(t *testing.T) {
	repo := repository.NewInMemoryRepository(fakeProjects)

	projects, err := repo.Fetch(context.Background(), 100)
	if err != nil {
		t.Fatal("expected no errors, but got one")
	}

	if !reflect.DeepEqual(projects, fakeProjects) {
		t.Errorf("expected %v, but got %v", fakeProjects, projects)
	}
}

func TestFetchWithLimit(t *testing.T) {
	repo := repository.NewInMemoryRepository(fakeProjects)
	want := fakeProjects[:5]

	projects, err := repo.Fetch(context.Background(), 5)
	if err != nil {
		t.Fatal("expected no errors, but got one")
	}

	if !reflect.DeepEqual(projects, want) {
		t.Errorf("expected %v, but got %v", want, projects)
	}
}

func TestFetchWithTimeout(t *testing.T) {
	repo := repository.NewInMemoryRepository(fakeProjects)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Nanosecond)
	defer cancel()

	_, err := repo.Fetch(ctx, 5)
	if err != context.DeadlineExceeded {
		t.Errorf("expected %v, but got %v", context.DeadlineExceeded, err)
	}
}
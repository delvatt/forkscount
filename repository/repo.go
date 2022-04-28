package repository

import (
	"context"
)

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ForksCount  int    `json:"forksCount"`
}

type Repository interface {
	Fetch(context.Context, int) ([]Project, error)
}

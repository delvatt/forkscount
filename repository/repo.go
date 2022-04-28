package repository

import "context"

type repository interface {
	fetch(context.Context, int)
}

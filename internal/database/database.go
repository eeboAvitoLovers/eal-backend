package database

import (
	"context"

	"github.com/eeboAvitoLovers/eal-backend/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller struct {
	Client *pgxpool.Pool
}

func (c *Controller) Create(ctx context.Context, data model.Message) error {
	return nil
}

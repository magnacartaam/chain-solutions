package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/repository"
)

type PostgresRepo struct {
	db *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) repository.Repository {
	return &PostgresRepo{
		db: db,
	}
}

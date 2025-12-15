package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
)

func (r *PostgresRepo) GetOpenBatch(ctx context.Context) (*domain.Batch, error) {
	query := `SELECT batch_id, status, created_at FROM batches WHERE status = 'OPEN' ORDER BY created_at ASC LIMIT 1`
	var b domain.Batch
	err := r.db.QueryRow(ctx, query).Scan(&b.BatchID, &b.Status, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *PostgresRepo) CreateBatch(ctx context.Context) (*domain.Batch, error) {
	query := `INSERT INTO batches (status) VALUES ('OPEN') RETURNING batch_id, status, created_at`
	var b domain.Batch
	err := r.db.QueryRow(ctx, query).Scan(&b.BatchID, &b.Status, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *PostgresRepo) AddSpinToBatch(ctx context.Context, spinID string, batchID int64) error {
	query := `UPDATE spins SET batch_id = $1 WHERE spin_id = $2`
	_, err := r.db.Exec(ctx, query, batchID, spinID)
	return err
}

func (r *PostgresRepo) CloseBatch(ctx context.Context, batchID int64, merkleRoot string, txSig string) error {
	query := `
		UPDATE batches 
		SET status = 'COMMITTED', merkle_root = $1, solana_tx_sig = $2, committed_at = NOW() 
		WHERE batch_id = $3
	`
	_, err := r.db.Exec(ctx, query, merkleRoot, txSig, batchID)
	return err
}

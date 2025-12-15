package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
)

func (r *PostgresRepo) CreateSpin(ctx context.Context, spin *domain.Spin) error {
	query := `
		INSERT INTO spins (
			spin_id, session_id, wallet_address, spin_nonce,
			server_seed, client_seed, server_seed_hash,
			bet_amount, payout_amount, outcome_json, leaf_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(ctx, query,
		spin.SpinID,
		spin.SessionID,
		spin.WalletAddress,
		spin.SpinNonce,
		spin.ServerSeed,
		spin.ClientSeed,
		spin.ServerSeedHash,
		spin.BetAmount,
		spin.PayoutAmount,
		spin.Outcome,
		spin.LeafHash,
	)
	return err
}

func (r *PostgresRepo) GetSpinsByWallet(ctx context.Context, walletAddress string, limit int, offset int) ([]domain.Spin, error) {
	query := `
		SELECT spin_id, session_id, wallet_address, spin_nonce,
		       server_seed, client_seed, server_seed_hash,
		       bet_amount, payout_amount, outcome_json, leaf_hash, batch_id, created_at
		FROM spins
		WHERE wallet_address = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, walletAddress, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spins []domain.Spin
	for rows.Next() {
		var s domain.Spin
		err := rows.Scan(
			&s.SpinID, &s.SessionID, &s.WalletAddress, &s.SpinNonce,
			&s.ServerSeed, &s.ClientSeed, &s.ServerSeedHash,
			&s.BetAmount, &s.PayoutAmount, &s.Outcome, &s.LeafHash, &s.BatchID, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		spins = append(spins, s)
	}
	return spins, nil
}

func (r *PostgresRepo) GetSpinCount(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM spins WHERE session_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, sessionID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *PostgresRepo) GetUnbatchedSpins(ctx context.Context, limit int) ([]domain.Spin, error) {
	query := `
			SELECT spin_id, leaf_hash 
			FROM spins 
			WHERE batch_id IS NULL 
			ORDER BY created_at 
			LIMIT $1
		`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spins []domain.Spin
	for rows.Next() {
		var s domain.Spin
		if err := rows.Scan(&s.SpinID, &s.LeafHash); err != nil {
			return nil, err
		}
		spins = append(spins, s)
	}
	return spins, nil
}

func (r *PostgresRepo) GetBatchSpins(ctx context.Context, batchID int64) ([]domain.Spin, error) {
	query := `SELECT spin_id, leaf_hash FROM spins WHERE batch_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.Query(ctx, query, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spins []domain.Spin
	for rows.Next() {
		var s domain.Spin
		if err := rows.Scan(&s.SpinID, &s.LeafHash); err != nil {
			return nil, err
		}
		spins = append(spins, s)
	}
	return spins, nil
}

func (r *PostgresRepo) GetBatch(ctx context.Context, batchID int64) (*domain.Batch, error) {
	query := `SELECT batch_id, status, merkle_root, solana_tx_sig FROM batches WHERE batch_id = $1`
	var b domain.Batch
	err := r.db.QueryRow(ctx, query, batchID).Scan(&b.BatchID, &b.Status, &b.MerkleRoot, &b.SolanaTxSig)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *PostgresRepo) GetSpin(ctx context.Context, spinID string) (*domain.Spin, error) {
	query := `
		SELECT 
			spin_id, session_id, wallet_address, spin_nonce,
			server_seed, client_seed, server_seed_hash,
			bet_amount, payout_amount, outcome_json, leaf_hash, batch_id, created_at
		FROM spins
		WHERE spin_id = $1
	`

	var s domain.Spin
	var outcomeBytes []byte

	err := r.db.QueryRow(ctx, query, spinID).Scan(
		&s.SpinID,
		&s.SessionID,
		&s.WalletAddress,
		&s.SpinNonce,
		&s.ServerSeed,
		&s.ClientSeed,
		&s.ServerSeedHash,
		&s.BetAmount,
		&s.PayoutAmount,
		&outcomeBytes,
		&s.LeafHash,
		&s.BatchID,
		&s.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(outcomeBytes, &s.Outcome); err != nil {
		return nil, fmt.Errorf("failed to parse outcome json: %w", err)
	}

	return &s, nil
}

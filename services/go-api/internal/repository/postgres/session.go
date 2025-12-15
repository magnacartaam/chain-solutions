package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
)

func (r *PostgresRepo) CreateSession(ctx context.Context, session *domain.Session) error {
	invalidateQuery := `UPDATE sessions SET is_active = FALSE WHERE wallet_address = $1`
	_, err := r.db.Exec(ctx, invalidateQuery, session.WalletAddress)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO sessions (
			session_id, wallet_address, playable_balance, 
			next_server_seed, next_server_seed_hash, is_active
		) VALUES ($1, $2, $3, $4, $5, TRUE)
	`
	_, err = r.db.Exec(ctx, query,
		session.SessionID,
		session.WalletAddress,
		session.PlayableBalance,
		session.NextServerSeed,
		session.NextServerSeedHash,
	)
	return err
}

func (r *PostgresRepo) GetActiveSession(ctx context.Context, walletAddress string) (*domain.Session, error) {
	query := `
		SELECT session_id, wallet_address, playable_balance, next_server_seed, next_server_seed_hash, is_active, created_at
		FROM sessions
		WHERE wallet_address = $1 AND is_active = TRUE
		LIMIT 1
	`
	var s domain.Session
	err := r.db.QueryRow(ctx, query, walletAddress).Scan(
		&s.SessionID,
		&s.WalletAddress,
		&s.PlayableBalance,
		&s.NextServerSeed,
		&s.NextServerSeedHash,
		&s.IsActive,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepo) GetLatestSession(ctx context.Context, walletAddress string) (*domain.Session, error) {
	query := `
		SELECT session_id, wallet_address, playable_balance, next_server_seed, next_server_seed_hash, is_active, created_at
		FROM sessions
		WHERE wallet_address = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var s domain.Session
	err := r.db.QueryRow(ctx, query, walletAddress).Scan(
		&s.SessionID,
		&s.WalletAddress,
		&s.PlayableBalance,
		&s.NextServerSeed,
		&s.NextServerSeedHash,
		&s.IsActive,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepo) UpdateSessionState(ctx context.Context, sessionID string, newBalance string, newSeed string, newHash string) error {
	query := `
		UPDATE sessions 
		SET playable_balance = $1, next_server_seed = $2, next_server_seed_hash = $3 
		WHERE session_id = $4
	`
	_, err := r.db.Exec(ctx, query, newBalance, newSeed, newHash, sessionID)
	return err
}

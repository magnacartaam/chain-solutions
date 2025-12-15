package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/domain"
	"github.com/shopspring/decimal"
)

func (r *PostgresRepo) CreateUser(ctx context.Context, walletAddress string) error {
	query := `
		INSERT INTO users (wallet_address, next_withdrawal_nonce)
		VALUES ($1, 1)
		ON CONFLICT (wallet_address) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, walletAddress)
	return err
}

func (r *PostgresRepo) GetUser(ctx context.Context, walletAddress string) (*domain.User, error) {
	query := `
		SELECT wallet_address, next_withdrawal_nonce, pending_withdrawal_amount, pending_withdrawal_signature, created_at
		FROM users
		WHERE wallet_address = $1
	`
	var user domain.User
	err := r.db.QueryRow(ctx, query, walletAddress).Scan(
		&user.WalletAddress,
		&user.NextWithdrawalNonce,
		&user.PendingWithdrawalAmount,
		&user.PendingWithdrawalSignature,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *PostgresRepo) SetPendingWithdrawal(ctx context.Context, walletAddress string, amount decimal.Decimal, signature string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	querySession := `
		UPDATE sessions 
		SET playable_balance = playable_balance - $1 
		WHERE wallet_address = $2 AND is_active = TRUE AND playable_balance >= $1
	`
	tag, err := tx.Exec(ctx, querySession, amount, walletAddress)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient funds or no active session")
	}

	queryUser := `
		UPDATE users 
		SET pending_withdrawal_amount = $1,
		    pending_withdrawal_signature = $2,
		    next_withdrawal_nonce = next_withdrawal_nonce + 1
		WHERE wallet_address = $3
	`
	_, err = tx.Exec(ctx, queryUser, amount, signature, walletAddress)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepo) CompleteWithdrawal(ctx context.Context, walletAddress string) error {
	query := `
		UPDATE users 
		SET pending_withdrawal_amount = 0,
		    pending_withdrawal_signature = ''
		WHERE wallet_address = $1
	`
	_, err := r.db.Exec(ctx, query, walletAddress)
	return err
}

func (r *PostgresRepo) RefundWithdrawal(ctx context.Context, walletAddress string, amount decimal.Decimal, correctNextNonce int64, fallbackSession *domain.Session) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queryUser := `
		UPDATE users 
		SET pending_withdrawal_amount = 0,
		    pending_withdrawal_signature = '',
		    next_withdrawal_nonce = $1
		WHERE wallet_address = $2
	`
	_, err = tx.Exec(ctx, queryUser, correctNextNonce, walletAddress)
	if err != nil {
		return err
	}

	querySession := `
		UPDATE sessions 
		SET playable_balance = playable_balance + $1 
		WHERE wallet_address = $2 AND is_active = TRUE
	`
	tag, err := tx.Exec(ctx, querySession, amount, walletAddress)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		_, err = tx.Exec(ctx, `UPDATE sessions SET is_active = FALSE WHERE wallet_address = $1`, walletAddress)
		if err != nil {
			return err
		}

		queryInsert := `
			INSERT INTO sessions (
				session_id, wallet_address, playable_balance, 
				next_server_seed, next_server_seed_hash, is_active
			) VALUES ($1, $2, $3, $4, $5, TRUE)
		`
		_, err = tx.Exec(ctx, queryInsert,
			fallbackSession.SessionID,
			fallbackSession.WalletAddress,
			amount,
			fallbackSession.NextServerSeed,
			fallbackSession.NextServerSeedHash,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepo) GetNextWithdrawalNonce(ctx context.Context, walletAddress string) (int64, error) {
	query := `SELECT next_withdrawal_nonce FROM users WHERE wallet_address = $1`
	var nonce int64
	err := r.db.QueryRow(ctx, query, walletAddress).Scan(&nonce)
	return nonce, err
}

func (r *PostgresRepo) IncrementWithdrawalNonce(ctx context.Context, walletAddress string) error {
	query := `
		UPDATE users 
		SET next_withdrawal_nonce = next_withdrawal_nonce + 1 
		WHERE wallet_address = $1
	`
	_, err := r.db.Exec(ctx, query, walletAddress)
	return err
}

func (r *PostgresRepo) CheckDepositProcessed(ctx context.Context, txSig string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM processed_deposits WHERE tx_sig = $1)`
	err := r.db.QueryRow(ctx, query, txSig).Scan(&exists)
	return exists, err
}

func (r *PostgresRepo) RecordDeposit(ctx context.Context, txSig string, walletAddress string, amount uint64, fallbackSession *domain.Session) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO users (wallet_address, next_withdrawal_nonce)
		VALUES ($1, 1)
		ON CONFLICT (wallet_address) DO NOTHING
	`, walletAddress)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO processed_deposits (tx_sig, wallet_address, amount_lamports)
		VALUES ($1, $2, $3)
	`, txSig, walletAddress, amount)
	if err != nil {
		return err
	}

	amountSol := decimal.NewFromInt(int64(amount)).Div(decimal.NewFromInt(1_000_000_000))

	queryUpdate := `
		UPDATE sessions 
		SET playable_balance = playable_balance + $1 
		WHERE wallet_address = $2 AND is_active = TRUE
	`
	tag, err := tx.Exec(ctx, queryUpdate, amountSol, walletAddress)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		queryInsert := `
			INSERT INTO sessions (
				session_id, wallet_address, playable_balance, 
				next_server_seed, next_server_seed_hash, is_active
			) VALUES ($1, $2, $3, $4, $5, TRUE)
		`
		_, err = tx.Exec(ctx, queryInsert,
			fallbackSession.SessionID,
			fallbackSession.WalletAddress,
			amountSol,
			fallbackSession.NextServerSeed,
			fallbackSession.NextServerSeedHash,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

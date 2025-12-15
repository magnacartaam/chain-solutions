CREATE TABLE users
(
    wallet_address        VARCHAR(44) PRIMARY KEY,
    next_withdrawal_nonce BIGINT                   DEFAULT 1 NOT NULL,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE sessions
(
    session_id            UUID PRIMARY KEY,
    wallet_address        VARCHAR(44)    NOT NULL REFERENCES users (wallet_address),
    playable_balance      DECIMAL(20, 9) NOT NULL  DEFAULT 0,

    next_server_seed      VARCHAR(64)    NOT NULL,
    next_server_seed_hash VARCHAR(64)    NOT NULL,

    is_active             BOOLEAN                  DEFAULT TRUE,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE spins
(
    spin_id          UUID PRIMARY KEY,
    session_id       UUID           NOT NULL REFERENCES sessions (session_id),
    wallet_address   VARCHAR(44)    NOT NULL REFERENCES users (wallet_address),

    spin_nonce       BIGINT         NOT NULL,

    server_seed      VARCHAR(64)    NOT NULL,
    client_seed      VARCHAR(64)    NOT NULL,
    server_seed_hash VARCHAR(64)    NOT NULL,

    bet_amount       DECIMAL(20, 9) NOT NULL,
    payout_amount    DECIMAL(20, 9) NOT NULL,
    outcome_json     JSONB          NOT NULL,

    leaf_hash        VARCHAR(64)    NOT NULL,
    batch_id         BIGINT,

    created_at       TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE batches
(
    batch_id      BIGSERIAL PRIMARY KEY,
    status        VARCHAR(20)              DEFAULT 'OPEN',
    merkle_root   VARCHAR(64),
    solana_tx_sig VARCHAR(88),
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    committed_at  TIMESTAMP WITH TIME ZONE
);

CREATE TABLE merkle_proofs
(
    spin_id      UUID PRIMARY KEY REFERENCES spins (spin_id),
    proof_hashes JSONB NOT NULL
);

CREATE TABLE processed_deposits (
                                    tx_sig VARCHAR(88) PRIMARY KEY,
                                    wallet_address VARCHAR(44) NOT NULL,
                                    amount_lamports BIGINT NOT NULL,
                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_spins_wallet ON spins (wallet_address);
CREATE INDEX idx_spins_batch ON spins (batch_id);
CREATE INDEX idx_batches_status ON batches (status);



ALTER TABLE users ADD COLUMN pending_withdrawal_amount DECIMAL(20, 9) NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN pending_withdrawal_signature VARCHAR(128) NOT NULL DEFAULT '';
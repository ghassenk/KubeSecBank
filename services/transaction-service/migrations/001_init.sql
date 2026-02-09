-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id              UUID PRIMARY KEY,
    from_account_id UUID NOT NULL,
    to_account_id   UUID NOT NULL,
    amount          DECIMAL(18, 2) NOT NULL CHECK (amount > 0),
    currency        VARCHAR(3) NOT NULL,
    type            VARCHAR(20) NOT NULL CHECK (type IN ('transfer', 'payment', 'deposit', 'withdrawal')),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'reversed')),
    description     TEXT DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for looking up transactions by either account (sender or receiver)
CREATE INDEX idx_transactions_from_account ON transactions (from_account_id, created_at DESC);
CREATE INDEX idx_transactions_to_account   ON transactions (to_account_id, created_at DESC);

-- Index for filtering by status
CREATE INDEX idx_transactions_status ON transactions (status);

-- Index for date-range queries
CREATE INDEX idx_transactions_created_at ON transactions (created_at DESC);

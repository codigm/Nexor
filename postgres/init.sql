-- OLYMPUS INITIALIZATION SQL (HERMES LAYER)
-- ============================================================================

-- 1. EXTENSIONS
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 2. ACCOUNTS TABLE
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank VARCHAR(50) NOT NULL,
    user_name VARCHAR(100) NOT NULL,
    account_number VARCHAR(50) UNIQUE NOT NULL,
    balance NUMERIC(20,4) NOT NULL DEFAULT 0.0000 CHECK (balance >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);



-- 4. JOURNAL ENTRIES TABLE (DOUBLE-ENTRY BOOKKEEPING)
CREATE TABLE journal_entries (
    journal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    debit_account VARCHAR(50) NOT NULL,
    credit_account VARCHAR(50) NOT NULL,
    amount NUMERIC(20,4) NOT NULL CHECK (amount > 0),
    currency CHAR(3) DEFAULT 'INR',
    entry_hash VARCHAR(64) NOT NULL,
    prev_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);



-- 6. OUTBOX EVENTS TABLE (Exactly-once delivery)
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

-- ============================================================================
-- INDEXES
-- ============================================================================
CREATE INDEX idx_accounts_bank ON accounts(bank);
CREATE INDEX idx_journal_transaction_id ON journal_entries(transaction_id);
CREATE INDEX idx_outbox_status ON outbox_events(status);

-- ============================================================================
-- VIEWS
-- ============================================================================
-- Trial Balance View to ensure debits and credits match perfectly
CREATE OR REPLACE VIEW trial_balance AS
SELECT 
  SUM(amount) as total_debits,
  SUM(amount) as total_credits,
  SUM(amount) - SUM(amount) as diff
FROM journal_entries;

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- AlphaBank Users (10 users, 1 Lakh each)
INSERT INTO accounts (bank, user_name, account_number, balance) VALUES
('AlphaBank', 'Alpha User 1', 'ALPHA001', 100000.0000),
('AlphaBank', 'Alpha User 2', 'ALPHA002', 100000.0000),
('AlphaBank', 'Alpha User 3', 'ALPHA003', 100000.0000),
('AlphaBank', 'Alpha User 4', 'ALPHA004', 100000.0000),
('AlphaBank', 'Alpha User 5', 'ALPHA005', 100000.0000),
('AlphaBank', 'Alpha User 6', 'ALPHA006', 100000.0000),
('AlphaBank', 'Alpha User 7', 'ALPHA007', 100000.0000),
('AlphaBank', 'Alpha User 8', 'ALPHA008', 100000.0000),
('AlphaBank', 'Alpha User 9', 'ALPHA009', 100000.0000),
('AlphaBank', 'Alpha User 10', 'ALPHA010', 100000.0000);

-- BetaBank Users (10 users, 1 Lakh each)
INSERT INTO accounts (bank, user_name, account_number, balance) VALUES
('BetaBank', 'Beta User 1', 'BETA001', 100000.0000),
('BetaBank', 'Beta User 2', 'BETA002', 100000.0000),
('BetaBank', 'Beta User 3', 'BETA003', 100000.0000),
('BetaBank', 'Beta User 4', 'BETA004', 100000.0000),
('BetaBank', 'Beta User 5', 'BETA005', 100000.0000),
('BetaBank', 'Beta User 6', 'BETA006', 100000.0000),
('BetaBank', 'Beta User 7', 'BETA007', 100000.0000),
('BetaBank', 'Beta User 8', 'BETA008', 100000.0000),
('BetaBank', 'Beta User 9', 'BETA009', 100000.0000),
('BetaBank', 'Beta User 10', 'BETA010', 100000.0000);

-- GammaBank Users (10 users, 1 Lakh each)
INSERT INTO accounts (bank, user_name, account_number, balance) VALUES
('GammaBank', 'Gamma User 1', 'GAMMA001', 100000.0000),
('GammaBank', 'Gamma User 2', 'GAMMA002', 100000.0000),
('GammaBank', 'Gamma User 3', 'GAMMA003', 100000.0000),
('GammaBank', 'Gamma User 4', 'GAMMA004', 100000.0000),
('GammaBank', 'Gamma User 5', 'GAMMA005', 100000.0000),
('GammaBank', 'Gamma User 6', 'GAMMA006', 100000.0000),
('GammaBank', 'Gamma User 7', 'GAMMA007', 100000.0000),
('GammaBank', 'Gamma User 8', 'GAMMA008', 100000.0000),
('GammaBank', 'Gamma User 9', 'GAMMA009', 100000.0000),
('GammaBank', 'Gamma User 10', 'GAMMA010', 100000.0000);

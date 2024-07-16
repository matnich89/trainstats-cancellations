CREATE TABLE IF NOT EXISTS cancellations
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    train_id VARCHAR(255) NOT NULL,
    operator VARCHAR(255) NOT NULL,
    cancellation_date DATE NOT NULL,
    cancellation_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cancellations_date ON cancellations(cancellation_date);
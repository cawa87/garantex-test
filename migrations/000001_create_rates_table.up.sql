-- Create rates table
CREATE TABLE IF NOT EXISTS rates (
    id BIGSERIAL PRIMARY KEY,
    ask DECIMAL(20, 8) NOT NULL,
    bid DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on timestamp for faster queries
CREATE INDEX IF NOT EXISTS idx_rates_timestamp ON rates(timestamp DESC);

-- Create index on created_at for faster queries
CREATE INDEX IF NOT EXISTS idx_rates_created_at ON rates(created_at DESC);

-- +migrate Up
CREATE TABLE IF NOT EXISTS agents (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    url VARCHAR(500) NOT NULL,
    api_key VARCHAR(255) NOT NULL UNIQUE,
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'offline' CHECK (status IN ('online', 'offline'))
);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_agents_api_key ON agents(api_key);
CREATE INDEX IF NOT EXISTS idx_agents_last_seen_at ON agents(last_seen_at);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);

-- +migrate Down
DROP TABLE IF EXISTS agents;
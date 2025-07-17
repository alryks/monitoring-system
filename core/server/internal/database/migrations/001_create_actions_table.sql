-- Создание таблицы actions
CREATE TABLE IF NOT EXISTS actions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    type varchar(100) NOT NULL,
    payload jsonb NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'pending',
    created timestamp NOT NULL DEFAULT now(),
    completed timestamp,
    response text,
    error text
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_actions_agent_id ON actions(agent_id);
CREATE INDEX IF NOT EXISTS idx_actions_status ON actions(status);
CREATE INDEX IF NOT EXISTS idx_actions_type ON actions(type);
CREATE INDEX IF NOT EXISTS idx_actions_created ON actions(created);
CREATE INDEX IF NOT EXISTS idx_actions_agent_status ON actions(agent_id, status);
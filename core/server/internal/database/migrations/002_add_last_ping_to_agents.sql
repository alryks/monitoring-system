-- Добавление колонки last_ping в таблицу agents
ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_ping timestamp;

-- Создание индекса для быстрого поиска по last_ping
CREATE INDEX IF NOT EXISTS idx_agents_last_ping ON agents(last_ping);
CREATE INDEX IF NOT EXISTS idx_agents_active_last_ping ON agents(is_active, last_ping); 
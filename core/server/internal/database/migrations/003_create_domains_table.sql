-- Создание таблицы доменов
CREATE TABLE IF NOT EXISTS domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    agent_ip VARCHAR(45) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    ssl_enabled BOOLEAN NOT NULL DEFAULT false,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание индексов для доменов
CREATE INDEX IF NOT EXISTS idx_domains_agent_id ON domains(agent_id);
CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name);
CREATE INDEX IF NOT EXISTS idx_domains_is_active ON domains(is_active);

-- Создание таблицы маршрутов доменов
CREATE TABLE IF NOT EXISTS domain_routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    container_name VARCHAR(255) NOT NULL,
    port VARCHAR(10) NOT NULL,
    path VARCHAR(255) NOT NULL DEFAULT '/',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание индексов для маршрутов
CREATE INDEX IF NOT EXISTS idx_domain_routes_domain_id ON domain_routes(domain_id);
CREATE INDEX IF NOT EXISTS idx_domain_routes_container_name ON domain_routes(container_name);
CREATE INDEX IF NOT EXISTS idx_domain_routes_is_active ON domain_routes(is_active);

-- Создание уникального индекса для комбинации domain_id и path
CREATE UNIQUE INDEX IF NOT EXISTS idx_domain_routes_unique ON domain_routes(domain_id, path);

-- Создание триггера для обновления updated timestamp
CREATE OR REPLACE FUNCTION update_updated_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Применение триггера к таблице domains
CREATE TRIGGER update_domains_updated BEFORE UPDATE ON domains
    FOR EACH ROW EXECUTE FUNCTION update_updated_column();

-- Применение триггера к таблице domain_routes
CREATE TRIGGER update_domain_routes_updated BEFORE UPDATE ON domain_routes
    FOR EACH ROW EXECUTE FUNCTION update_updated_column(); 
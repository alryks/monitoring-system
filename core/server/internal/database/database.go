package database

import (
	"database/sql"
	"fmt"
	"log"
)

// Connect подключается к базе данных PostgreSQL
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	// Создаем таблицы если их нет
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Выполняем миграции
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// createTables создает необходимые таблицы в базе данных
func createTables(db *sql.DB) error {
	queries := []string{
		// Включаем расширение для UUID
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,

		// Таблица пользователей
		`CREATE TABLE IF NOT EXISTS users (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			username varchar(255) NOT NULL UNIQUE,
			password_hash varchar(255) NOT NULL,
			email varchar(255),
			is_active boolean NOT NULL DEFAULT true,
			role varchar(50) NOT NULL DEFAULT 'user',
			created timestamp NOT NULL DEFAULT now(),
			last_login timestamp
		);`,

		// Индексы для users
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		`CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);`,

		// Таблица агентов
		`CREATE TABLE IF NOT EXISTS agents (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			name varchar(255) NOT NULL,
			token varchar(255) NOT NULL UNIQUE,
			is_active boolean NOT NULL DEFAULT true,
			created timestamp NOT NULL DEFAULT now(),
			last_ping timestamp
		);`,

		// Индексы для agents
		`CREATE INDEX IF NOT EXISTS idx_agents_token ON agents(token);`,
		`CREATE INDEX IF NOT EXISTS idx_agents_is_active ON agents(is_active);`,

		// Таблица пингов агентов
		`CREATE TABLE IF NOT EXISTS agent_pings (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			created timestamp NOT NULL DEFAULT now()
		);`,

		// Индексы для agent_pings
		`CREATE INDEX IF NOT EXISTS idx_agent_pings_agent_id ON agent_pings(agent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_agent_pings_created ON agent_pings(created);`,
		`CREATE INDEX IF NOT EXISTS idx_agent_pings_agent_created ON agent_pings(agent_id, created);`,

		// Таблица метрик CPU
		`CREATE TABLE IF NOT EXISTS cpu_metrics (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			ping_id uuid NOT NULL REFERENCES agent_pings(id) ON DELETE CASCADE,
			cpu_name varchar(50) NOT NULL,
			usage_percent decimal(5,4) NOT NULL
		);`,

		// Индексы для cpu_metrics
		`CREATE INDEX IF NOT EXISTS idx_cpu_metrics_ping_id ON cpu_metrics(ping_id);`,
		`CREATE INDEX IF NOT EXISTS idx_cpu_metrics_cpu_name ON cpu_metrics(cpu_name);`,

		// Таблица метрик памяти
		`CREATE TABLE IF NOT EXISTS memory_metrics (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			ping_id uuid NOT NULL REFERENCES agent_pings(id) ON DELETE CASCADE,
			ram_total_mb bigint NOT NULL,
			ram_usage_mb bigint NOT NULL,
			swap_total_mb bigint NOT NULL,
			swap_usage_mb bigint NOT NULL
		);`,

		// Индексы для memory_metrics
		`CREATE INDEX IF NOT EXISTS idx_memory_metrics_ping_id ON memory_metrics(ping_id);`,

		// Таблица метрик дисков
		`CREATE TABLE IF NOT EXISTS disk_metrics (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			ping_id uuid NOT NULL REFERENCES agent_pings(id) ON DELETE CASCADE,
			disk_name varchar(50) NOT NULL,
			read_bytes bigint NOT NULL,
			write_bytes bigint NOT NULL,
			reads bigint NOT NULL,
			writes bigint NOT NULL
		);`,

		// Индексы для disk_metrics
		`CREATE INDEX IF NOT EXISTS idx_disk_metrics_ping_id ON disk_metrics(ping_id);`,
		`CREATE INDEX IF NOT EXISTS idx_disk_metrics_disk_name ON disk_metrics(disk_name);`,

		// Таблица метрик сети
		`CREATE TABLE IF NOT EXISTS network_metrics (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			ping_id uuid NOT NULL REFERENCES agent_pings(id) ON DELETE CASCADE,
			public_ip inet NOT NULL,
			sent_bytes bigint NOT NULL,
			received_bytes bigint NOT NULL
		);`,

		// Индексы для network_metrics
		`CREATE INDEX IF NOT EXISTS idx_network_metrics_ping_id ON network_metrics(ping_id);`,
		`CREATE INDEX IF NOT EXISTS idx_network_metrics_public_ip ON network_metrics(public_ip);`,

		// Таблица контейнеров
		`CREATE TABLE IF NOT EXISTS containers (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			ping_id uuid NOT NULL REFERENCES agent_pings(id) ON DELETE CASCADE,
			container_id varchar(64) NOT NULL,
			name varchar(255) NOT NULL,
			image_id varchar(64) NOT NULL,
			status varchar(50) NOT NULL,
			restart_count integer NOT NULL DEFAULT 0,
			created_at timestamp NOT NULL,
			ip_address inet,
			mac_address varchar(17),
			cpu_usage_percent decimal(8,6),
			memory_usage_mb bigint,
			network_sent_bytes bigint,
			network_received_bytes bigint
		);`,

		// Индексы для containers
		`CREATE INDEX IF NOT EXISTS idx_containers_ping_id ON containers(ping_id);`,
		`CREATE INDEX IF NOT EXISTS idx_containers_container_id ON containers(container_id);`,
		`CREATE INDEX IF NOT EXISTS idx_containers_name ON containers(name);`,
		`CREATE INDEX IF NOT EXISTS idx_containers_status ON containers(status);`,

		// Таблица логов контейнеров
		`CREATE TABLE IF NOT EXISTS container_logs (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			container_id uuid NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
			log_line text NOT NULL,
			timestamp timestamp NOT NULL DEFAULT now()
		);`,

		// Индексы для container_logs
		`CREATE INDEX IF NOT EXISTS idx_container_logs_container_id ON container_logs(container_id);`,
		`CREATE INDEX IF NOT EXISTS idx_container_logs_timestamp ON container_logs(timestamp);`,

		// Таблица образов
		`CREATE TABLE IF NOT EXISTS images (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			ping_id uuid NOT NULL REFERENCES agent_pings(id) ON DELETE CASCADE,
			image_id varchar(64) NOT NULL,
			created timestamp NOT NULL,
			size bigint NOT NULL,
			architecture varchar(50) NOT NULL DEFAULT 'amd64'
		);`,

		// Индексы для images
		`CREATE INDEX IF NOT EXISTS idx_images_ping_id ON images(ping_id);`,
		`CREATE INDEX IF NOT EXISTS idx_images_image_id ON images(image_id);`,

		// Таблица тегов образов
		`CREATE TABLE IF NOT EXISTS image_tags (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			image_id uuid NOT NULL REFERENCES images(id) ON DELETE CASCADE,
			tag varchar(255) NOT NULL
		);`,

		// Индексы для image_tags
		`CREATE INDEX IF NOT EXISTS idx_image_tags_image_id ON image_tags(image_id);`,
		`CREATE INDEX IF NOT EXISTS idx_image_tags_tag ON image_tags(tag);`,

		// Таблица действий
		`CREATE TABLE IF NOT EXISTS actions (
			id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			type varchar(100) NOT NULL,
			payload jsonb NOT NULL,
			status varchar(20) NOT NULL DEFAULT 'pending',
			created timestamp NOT NULL DEFAULT now(),
			completed timestamp,
			response text,
			error text
		);`,

		// Индексы для actions
		`CREATE INDEX IF NOT EXISTS idx_actions_agent_id ON actions(agent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_actions_status ON actions(status);`,
		`CREATE INDEX IF NOT EXISTS idx_actions_created ON actions(created DESC);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
		}
	}

	log.Println("Database tables created successfully")
	return nil
}

// runMigrations выполняет миграции базы данных
func runMigrations(db *sql.DB) error {
	migrations := []string{
		// Миграция 002: добавление колонки last_ping в таблицу agents
		`ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_ping timestamp;`,
		`CREATE INDEX IF NOT EXISTS idx_agents_last_ping ON agents(last_ping);`,
		`CREATE INDEX IF NOT EXISTS idx_agents_active_last_ping ON agents(is_active, last_ping);`,
		
		// Миграция 003: создание таблиц доменов
		`CREATE TABLE IF NOT EXISTS domains (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL UNIQUE,
			agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			agent_ip VARCHAR(45) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true,
			ssl_enabled BOOLEAN NOT NULL DEFAULT false,
			created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_domains_agent_id ON domains(agent_id);`,
		`CREATE INDEX IF NOT EXISTS idx_domains_name ON domains(name);`,
		`CREATE INDEX IF NOT EXISTS idx_domains_is_active ON domains(is_active);`,
		
		`CREATE TABLE IF NOT EXISTS domain_routes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
			container_name VARCHAR(255) NOT NULL,
			port VARCHAR(10) NOT NULL,
			path VARCHAR(255) NOT NULL DEFAULT '/',
			is_active BOOLEAN NOT NULL DEFAULT true,
			created TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_domain_routes_domain_id ON domain_routes(domain_id);`,
		`CREATE INDEX IF NOT EXISTS idx_domain_routes_container_name ON domain_routes(container_name);`,
		`CREATE INDEX IF NOT EXISTS idx_domain_routes_is_active ON domain_routes(is_active);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_domain_routes_unique ON domain_routes(domain_id, path);`,
		
		// Создание триггера для обновления updated timestamp
		`CREATE OR REPLACE FUNCTION update_updated_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';`,
		
		`CREATE TRIGGER update_domains_updated BEFORE UPDATE ON domains
			FOR EACH ROW EXECUTE FUNCTION update_updated_column();`,
		
		`CREATE TRIGGER update_domain_routes_updated BEFORE UPDATE ON domain_routes
			FOR EACH ROW EXECUTE FUNCTION update_updated_column();`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration: %s, error: %w", migration, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}

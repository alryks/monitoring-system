package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"monitoring-system/core/server/pkg/database"
	"monitoring-system/core/server/pkg/models"
)

type AgentService struct {
	db *database.DB
}

func NewAgentService(db *database.DB) *AgentService {
	return &AgentService{db: db}
}

func (s *AgentService) RegisterAgent(req models.RegisterRequest) (*models.RegisterResponse, error) {
	// Generate API key
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Insert agent into database
	query := `
		INSERT INTO agents (name, url, api_key, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	var agentID int
	err = s.db.QueryRow(query, req.Name, req.URL, apiKey).Scan(&agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert agent: %w", err)
	}

	return &models.RegisterResponse{
		ID:     agentID,
		Name:   req.Name,
		APIKey: apiKey,
	}, nil
}

func (s *AgentService) GetAgentByAPIKey(apiKey string) (*models.Agent, error) {
	query := `
		SELECT id, name, url, api_key, last_seen_at, created_at, updated_at, status
		FROM agents WHERE api_key = $1
	`

	agent := &models.Agent{}
	err := s.db.QueryRow(query, apiKey).Scan(
		&agent.ID, &agent.Name, &agent.URL, &agent.APIKey,
		&agent.LastSeenAt, &agent.CreatedAt, &agent.UpdatedAt, &agent.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	return agent, nil
}

func (s *AgentService) UpdateLastSeen(apiKey string) error {
	query := `
		UPDATE agents 
		SET last_seen_at = CURRENT_TIMESTAMP, 
		    updated_at = CURRENT_TIMESTAMP,
		    status = 'online'
		WHERE api_key = $1
	`

	_, err := s.db.Exec(query, apiKey)
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	return nil
}

func (s *AgentService) GetAllAgents() ([]models.Agent, error) {
	query := `
		SELECT id, name, url, last_seen_at, created_at, updated_at, status
		FROM agents
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents: %w", err)
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		var agent models.Agent
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.URL,
			&agent.LastSeenAt, &agent.CreatedAt, &agent.UpdatedAt, &agent.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		// Determine status based on last_seen_at
		if time.Since(agent.LastSeenAt) > 30*time.Second {
			agent.Status = "offline"
		} else {
			agent.Status = "online"
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

func (s *AgentService) generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

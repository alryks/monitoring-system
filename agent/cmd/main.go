package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/client"
)

type Agent struct {
	ID       string
	CoreURL  string
	APIKey   string
	Interval time.Duration
	docker   *client.Client
}

type HeartbeatRequest struct {
	AgentID   string    `json:"agent_id"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

type HeartbeatResponse struct {
	Status string `json:"status"`
	Tasks  []Task `json:"tasks"`
}

type Task struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

func NewAgent() *Agent {
	coreURL := os.Getenv("CORE_API_URL")
	if coreURL == "" {
		coreURL = "http://core:80/api"
	}

	agentID := os.Getenv("AGENT_ID")
	if agentID == "" {
		agentID = "agent-1"
	}

	intervalStr := os.Getenv("HEARTBEAT_INTERVAL")
	interval := 5 * time.Second
	if intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			interval = d
		}
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Printf("Warning: Failed to create Docker client: %v", err)
	}

	return &Agent{
		ID:       agentID,
		CoreURL:  coreURL,
		Interval: interval,
		docker:   dockerClient,
	}
}

func (a *Agent) Start() {
	log.Printf("Starting agent %s", a.ID)
	log.Printf("Core API URL: %s", a.CoreURL)
	log.Printf("Heartbeat interval: %v", a.Interval)

	// Test connection to Core
	if err := a.testConnection(); err != nil {
		log.Printf("Warning: Failed to connect to Core: %v", err)
	} else {
		log.Printf("Successfully connected to Core")
	}

	// Start heartbeat loop
	ticker := time.NewTicker(a.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.sendHeartbeat(); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
		}
	}
}

func (a *Agent) testConnection() error {
	url := fmt.Sprintf("%s/ping", a.CoreURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (a *Agent) sendHeartbeat() error {
	heartbeat := HeartbeatRequest{
		AgentID:   a.ID,
		Timestamp: time.Now(),
		Status:    "online",
	}

	body, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %w", err)
	}

	url := fmt.Sprintf("%s/heartbeat", a.CoreURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if a.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed with status %d: %s", resp.StatusCode, string(body))
	}

	var heartbeatResp HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&heartbeatResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("Heartbeat sent successfully. Status: %s, Tasks: %d", heartbeatResp.Status, len(heartbeatResp.Tasks))

	// Process tasks
	for _, task := range heartbeatResp.Tasks {
		if err := a.processTask(task); err != nil {
			log.Printf("Failed to process task %s: %v", task.ID, err)
		}
	}

	return nil
}

func (a *Agent) processTask(task Task) error {
	log.Printf("Processing task %s of type %s", task.ID, task.Type)

	switch task.Type {
	case "UPDATE_NGINX":
		log.Printf("Would update NGINX configuration: %+v", task.Payload)
	case "GET_LOGS":
		log.Printf("Would get logs: %+v", task.Payload)
	case "ISSUE_SSL":
		log.Printf("Would issue SSL certificate: %+v", task.Payload)
	default:
		log.Printf("Unknown task type: %s", task.Type)
	}

	return nil
}

func main() {
	agent := NewAgent()
	agent.Start()
}

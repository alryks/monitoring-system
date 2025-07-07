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
	"os/exec"
	"strings"
	"time"
)

type Agent struct {
	ID       string
	CoreURL  string
	APIKey   string
	Interval time.Duration
}

type RegisterRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type RegisterResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	APIKey string `json:"api_key"`
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
		log.Fatal("AGENT_ID environment variable is required")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	intervalStr := os.Getenv("HEARTBEAT_INTERVAL")
	interval := 5 * time.Second
	if intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			interval = d
		}
	}

	return &Agent{
		ID:       agentID,
		CoreURL:  coreURL,
		APIKey:   apiKey,
		Interval: interval,
	}
}

func (a *Agent) Start() {
	log.Printf("Starting agent %s", a.ID)
	log.Printf("Core API URL: %s", a.CoreURL)
	log.Printf("Heartbeat interval: %v", a.Interval)

	if err := a.testDockerConnection(); err != nil {
		log.Printf("Warning: Failed to connect to Docker: %v", err)
	} else {
		log.Printf("Successfully connected to Docker")
	}

	if err := a.testConnection(); err != nil {
		log.Printf("Warning: Failed to connect to Core: %v", err)
	} else {
		log.Printf("Successfully connected to Core")
	}

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

func (a *Agent) testDockerConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get docker version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	log.Printf("Docker version: %s", version)
	return nil
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
	req.Header.Set("Authorization", "Bearer "+a.APIKey)

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
		return a.processUpdateNginx(task)
	case "GET_LOGS":
		return a.processGetLogs(task)
	case "ISSUE_SSL":
		return a.processIssueSSL(task)
	case "DOCKER_COMMAND":
		return a.processDockerCommand(task)
	default:
		log.Printf("Unknown task type: %s", task.Type)
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (a *Agent) processUpdateNginx(task Task) error {
	log.Printf("Would update NGINX configuration: %+v", task.Payload)
	// TODO: NGINX update
	return nil
}

func (a *Agent) processGetLogs(task Task) error {
	containerID, ok := task.Payload["container_id"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid container_id in task payload")
	}

	tail := 100
	if t, ok := task.Payload["tail"].(float64); ok {
		tail = int(t)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "logs", "--tail", fmt.Sprintf("%d", tail), containerID)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}

	log.Printf("Container %s logs (%d lines):\n%s", containerID, tail, string(output))
	return nil
}

func (a *Agent) processIssueSSL(task Task) error {
	domain, ok := task.Payload["domain"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid domain in task payload")
	}

	log.Printf("Would issue SSL certificate for domain: %s", domain)
	// TODO: SSL certificate with acme.sh
	return nil
}

func (a *Agent) processDockerCommand(task Task) error {
	command, ok := task.Payload["command"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid command in task payload")
	}

	args := strings.Fields(command)
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	allowedCommands := []string{"ps", "stats", "logs", "inspect", "version"}
	if !contains(allowedCommands, args[0]) {
		return fmt.Errorf("command not allowed: %s", args[0])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fullArgs := append([]string{args[0]}, args[1:]...)
	cmd := exec.CommandContext(ctx, "docker", fullArgs...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute docker command: %w", err)
	}

	log.Printf("Docker command '%s' output:\n%s", command, string(output))
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func main() {
	agent := NewAgent()
	agent.Start()
}

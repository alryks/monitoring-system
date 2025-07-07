package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"monitoring-system/core/server/pkg/models"
	"monitoring-system/core/server/pkg/services"
)

type AgentHandler struct {
	agentService *services.AgentService
}

func NewAgentHandler(agentService *services.AgentService) *AgentHandler {
	return &AgentHandler{agentService: agentService}
}

// CreateNode создает новый узел администратором (вместо автоматической регистрации)
func (h *AgentHandler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Description == "" {
		req.Description = "Узел " + req.Name
	}

	resp, err := h.agentService.CreateNode(req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Node with this name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create node", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AgentHandler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	resp, err := h.agentService.RegisterAgent(req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Agent with this name already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to register agent", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AgentHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	// Extract API key from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	// Extract token from "Bearer <token>"
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
		return
	}
	apiKey := tokenParts[1]

	// Verify agent exists
	_, err := h.agentService.GetAgentByAPIKey(apiKey)
	if err != nil {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	// Parse heartbeat request
	var req models.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update last seen timestamp
	if err := h.agentService.UpdateLastSeen(apiKey); err != nil {
		http.Error(w, "Failed to update agent status", http.StatusInternalServerError)
		return
	}

	// Return response with empty tasks for now
	resp := models.HeartbeatResponse{
		Status: "ok",
		Tasks:  []models.Task{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AgentHandler) GetAgents(w http.ResponseWriter, r *http.Request) {
	agents, err := h.agentService.GetAllAgents()
	if err != nil {
		http.Error(w, "Failed to get agents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

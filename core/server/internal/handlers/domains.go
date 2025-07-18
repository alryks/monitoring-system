package handlers

import (
	"encoding/json"
	"net/http"

	"monitoring-system/core/server/internal/domains"
	"monitoring-system/core/server/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DomainHandler struct {
	domainService *domains.Service
}

func NewDomainHandler(domainService *domains.Service) *DomainHandler {
	return &DomainHandler{
		domainService: domainService,
	}
}

// CreateDomain создает новый домен
// @Summary Создать новый домен
// @Description Создает новый домен с указанным агентом
// @Tags domains
// @Accept json
// @Produce json
// @Param domain body models.CreateDomainRequest true "Данные домена"
// @Success 201 {object} models.Domain
// @Failure 400 
// @Failure 500 
// @Router /api/domains [post]
func (h *DomainHandler) CreateDomain(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	domain, err := h.domainService.CreateDomain(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(domain)
}

// GetDomains получает список всех доменов
// @Summary Получить список доменов
// @Description Возвращает список всех доменов с их маршрутами
// @Tags domains
// @Produce json
// @Success 200 {object} models.DomainListResponse``
// @Failure 500 
// @Router /api/domains [get]
func (h *DomainHandler) GetDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := h.domainService.GetDomains()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.DomainListResponse{
		Domains: make([]models.DomainDetail, len(domains)),
		Total:   len(domains),
	}

	for i, domain := range domains {
		response.Domains[i] = *domain
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDomain получает домен по ID
// @Summary Получить домен по ID
// @Description Возвращает детальную информацию о домене
// @Tags domains
// @Produce json
// @Param id path string true "ID домена"
// @Success 200 {object} models.DomainDetail
// @Failure 404 
// @Failure 500 
// @Router /api/domains/{id} [get]
func (h *DomainHandler) GetDomain(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid domain ID", http.StatusBadRequest)
		return
	}

	domain, err := h.domainService.GetDomainByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain)
}

// UpdateDomain обновляет домен
// @Summary Обновить домен
// @Description Обновляет информацию о домене
// @Tags domains
// @Accept json
// @Produce json
// @Param id path string true "ID домена"
// @Param domain body models.UpdateDomainRequest true "Данные для обновления"
// @Success 200 {object} models.Domain
// @Failure 400 
// @Failure 404 
// @Failure 500 
// @Router /api/domains/{id} [put]
func (h *DomainHandler) UpdateDomain(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid domain ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	domain, err := h.domainService.UpdateDomain(id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain)
}

// DeleteDomain удаляет домен
// @Summary Удалить домен
// @Description Удаляет домен и все его маршруты
// @Tags domains
// @Param id path string true "ID домена"
// @Success 204 "No Content"
// @Failure 400 
// @Failure 500 
// @Router /api/domains/{id} [delete]
func (h *DomainHandler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid domain ID", http.StatusBadRequest)
		return
	}

	err = h.domainService.DeleteDomain(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateDomainRoute создает маршрут для домена
// @Summary Создать маршрут домена
// @Description Создает новый маршрут для домена
// @Tags domain-routes
// @Accept json
// @Produce json
// @Param route body models.CreateDomainRouteRequest true "Данные маршрута"
// @Success 201 {object} models.DomainRoute
// @Failure 400 
// @Failure 500 
// @Router /api/domains/routes [post]
func (h *DomainHandler) CreateDomainRoute(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDomainRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	route, err := h.domainService.CreateDomainRoute(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(route)
}

// GetDomainRoutes получает маршруты домена
// @Summary Получить маршруты домена
// @Description Возвращает все маршруты для указанного домена
// @Tags domain-routes
// @Produce json
// @Param domain_id path string true "ID домена"
// @Success 200 {object} models.DomainRouteListResponse
// @Failure 400 
// @Failure 500 
// @Router /api/domains/{domain_id}/routes [get]
func (h *DomainHandler) GetDomainRoutes(w http.ResponseWriter, r *http.Request) {
	domainIDStr := chi.URLParam(r, "domain_id")
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, "Invalid domain ID", http.StatusBadRequest)
		return
	}

	routes, err := h.domainService.GetDomainRoutes(domainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.DomainRouteListResponse{
		Routes: routes,
		Total:  len(routes),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateDomainRoute обновляет маршрут домена
// @Summary Обновить маршрут домена
// @Description Обновляет информацию о маршруте домена
// @Tags domain-routes
// @Accept json
// @Produce json
// @Param id path string true "ID маршрута"
// @Param route body models.UpdateDomainRouteRequest true "Данные для обновления"
// @Success 200 {object} models.DomainRoute
// @Failure 400 
// @Failure 500 
// @Router /api/domains/routes/{id} [put]
func (h *DomainHandler) UpdateDomainRoute(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateDomainRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	route, err := h.domainService.UpdateDomainRoute(id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(route)
}

// DeleteDomainRoute удаляет маршрут домена
// @Summary Удалить маршрут домена
// @Description Удаляет маршрут домена
// @Tags domain-routes
// @Param id path string true "ID маршрута"
// @Success 204 "No Content"
// @Failure 400 
// @Failure 500 
// @Router /api/domains/routes/{id} [delete]
func (h *DomainHandler) DeleteDomainRoute(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}

	err = h.domainService.DeleteDomainRoute(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetAgentNginxConfig получает конфигурацию nginx для агента
// @Summary Получить конфигурацию nginx агента
// @Description Возвращает конфигурацию nginx для указанного агента
// @Tags nginx
// @Produce json
// @Param agent_id path string true "ID агента"
// @Success 200 {object} models.AgentNginxConfig
// @Failure 400 
// @Failure 500 
// @Router /api/agents/{agent_id}/nginx-config [get]
func (h *DomainHandler) GetAgentNginxConfig(w http.ResponseWriter, r *http.Request) {
	agentIDStr := chi.URLParam(r, "agent_id")
	agentID, err := uuid.Parse(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	config, err := h.domainService.GetAgentNginxConfig(agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// GetDomainStatus получает статус домена
// @Summary Получить статус домена
// @Description Возвращает статус домена с информацией о контейнерах
// @Tags domains
// @Produce json
// @Param id path string true "ID домена"
// @Success 200 {object} models.DomainStatus
// @Failure 400 
// @Failure 500 
// @Router /api/domains/{id}/status [get]
func (h *DomainHandler) GetDomainStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid domain ID", http.StatusBadRequest)
		return
	}

	status, err := h.domainService.GetDomainStatus(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

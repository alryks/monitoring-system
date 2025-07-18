package domains

import (
	"database/sql"
	"fmt"
	"time"

	"monitoring-system/core/server/internal/models"

	"github.com/google/uuid"
)

type Service struct {
	db *sql.DB
}

// createAction создает действие для агента
func (s *Service) createAction(agentID uuid.UUID, actionType string, payload map[string]interface{}) error {
	_, err := s.db.Exec(`
		INSERT INTO actions (id, agent_id, type, payload, status, created)
		VALUES (gen_random_uuid(), $1, $2, $3, 'pending', NOW())
	`, agentID, actionType, payload)

	if err != nil {
		return fmt.Errorf("failed to create action: %v", err)
	}

	return nil
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// CreateDomain создает новый домен
func (s *Service) CreateDomain(req *models.CreateDomainRequest) (*models.Domain, error) {
	// Получаем IP агента из последнего ping
	var agentIP string
	err := s.db.QueryRow(`
		SELECT nm.public_ip 
		FROM agents a
		JOIN agent_pings ap ON a.id = ap.agent_id
		JOIN network_metrics nm ON ap.id = nm.ping_id
		WHERE a.id = $1
		ORDER BY ap.created DESC
		LIMIT 1
	`, req.AgentID).Scan(&agentIP)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent IP: %v", err)
	}

	domain := &models.Domain{
		ID:         uuid.New(),
		Name:       req.Name,
		AgentID:    req.AgentID,
		AgentIP:    agentIP,
		IsActive:   true,
		SSLEnabled: req.SSLEnabled,
		Created:    time.Now(),
		Updated:    time.Now(),
	}

	_, err = s.db.Exec(`
		INSERT INTO domains (id, name, agent_id, agent_ip, is_active, ssl_enabled, created, updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, domain.ID, domain.Name, domain.AgentID, domain.AgentIP, domain.IsActive, domain.SSLEnabled, domain.Created, domain.Updated)

	if err != nil {
		return nil, fmt.Errorf("failed to create domain: %v", err)
	}

	return domain, nil
}

// GetDomainByID получает домен по ID
func (s *Service) GetDomainByID(id uuid.UUID) (*models.DomainDetail, error) {
	domain := &models.DomainDetail{}

	err := s.db.QueryRow(`
		SELECT d.id, d.name, d.agent_id, d.agent_ip, d.is_active, d.ssl_enabled, d.created, d.updated,
		       a.name as agent_name
		FROM domains d
		JOIN agents a ON d.agent_id = a.id
		WHERE d.id = $1
	`, id).Scan(
		&domain.ID, &domain.Name, &domain.AgentID, &domain.AgentIP, &domain.IsActive, &domain.SSLEnabled, &domain.Created, &domain.Updated,
		&domain.AgentName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %v", err)
	}

	// Получаем маршруты домена
	routes, err := s.GetDomainRoutes(domain.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain routes: %v", err)
	}
	domain.Routes = routes

	return domain, nil
}

// GetDomains получает список всех доменов
func (s *Service) GetDomains() ([]*models.DomainDetail, error) {
	rows, err := s.db.Query(`
		SELECT d.id, d.name, d.agent_id, d.agent_ip, d.is_active, d.ssl_enabled, d.created, d.updated,
		       a.name as agent_name
		FROM domains d
		JOIN agents a ON d.agent_id = a.id
		ORDER BY d.created DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get domains: %v", err)
	}
	defer rows.Close()

	var domains []*models.DomainDetail
	for rows.Next() {
		domain := &models.DomainDetail{}
		err := rows.Scan(
			&domain.ID, &domain.Name, &domain.AgentID, &domain.AgentIP, &domain.IsActive, &domain.SSLEnabled, &domain.Created, &domain.Updated,
			&domain.AgentName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain: %v", err)
		}

		// Получаем маршруты домена
		routes, err := s.GetDomainRoutes(domain.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get domain routes: %v", err)
		}
		if routes == nil {
			domain.Routes = []models.DomainRoute{}
		} else {
			domain.Routes = routes
		}

		domains = append(domains, domain)
	}

	return domains, nil
}

// UpdateDomain обновляет домен
func (s *Service) UpdateDomain(id uuid.UUID, req *models.UpdateDomainRequest) (*models.Domain, error) {
	// Получаем текущий домен
	domain, err := s.GetDomainByID(id)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	if req.Name != nil {
		domain.Name = *req.Name
	}
	if req.AgentID != nil {
		domain.AgentID = *req.AgentID
		// Обновляем IP агента из последнего ping
		var agentIP string
		err := s.db.QueryRow(`
			SELECT nm.public_ip 
			FROM agents a
			JOIN agent_pings ap ON a.id = ap.agent_id
			JOIN network_metrics nm ON ap.id = nm.ping_id
			WHERE a.id = $1
			ORDER BY ap.created DESC
			LIMIT 1
		`, domain.AgentID).Scan(&agentIP)
		if err != nil {
			return nil, fmt.Errorf("failed to get agent IP: %v", err)
		}
		domain.AgentIP = agentIP
	}
	if req.IsActive != nil {
		domain.IsActive = *req.IsActive
	}
	if req.SSLEnabled != nil {
		domain.SSLEnabled = *req.SSLEnabled
	}

	domain.Updated = time.Now()

	_, err = s.db.Exec(`
		UPDATE domains 
		SET name = $1, agent_id = $2, agent_ip = $3, is_active = $4, ssl_enabled = $5, updated = $6
		WHERE id = $7
	`, domain.Name, domain.AgentID, domain.AgentIP, domain.IsActive, domain.SSLEnabled, domain.Updated, id)

	if err != nil {
		return nil, fmt.Errorf("failed to update domain: %v", err)
	}

	// Создаем действие для обновления nginx конфигурации при изменении SSL
	if req.SSLEnabled != nil {
		if err := s.createNginxUpdateAction(id); err != nil {
			// Логируем ошибку, но не прерываем обновление домена
			fmt.Printf("Warning: failed to create nginx update action: %v\n", err)
		}
	}

	return &domain.Domain, nil
}

// DeleteDomain удаляет домен
func (s *Service) DeleteDomain(id uuid.UUID) error {
	_, err := s.db.Exec("DELETE FROM domains WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %v", err)
	}
	return nil
}

// CreateDomainRoute создает маршрут для домена
func (s *Service) CreateDomainRoute(req *models.CreateDomainRouteRequest) (*models.DomainRoute, error) {
	route := &models.DomainRoute{
		ID:            uuid.New(),
		DomainID:      req.DomainID,
		ContainerName: req.ContainerName,
		Port:          req.Port,
		Path:          req.Path,
		IsActive:      true,
		Created:       time.Now(),
		Updated:       time.Now(),
	}

	if route.Path == "" {
		route.Path = "/"
	}

	_, err := s.db.Exec(`
		INSERT INTO domain_routes (id, domain_id, container_name, port, path, is_active, created, updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, route.ID, route.DomainID, route.ContainerName, route.Port, route.Path, route.IsActive, route.Created, route.Updated)

	if err != nil {
		return nil, fmt.Errorf("failed to create domain route: %v", err)
	}

	// Создаем действие для обновления nginx конфигурации
	if err := s.createNginxUpdateAction(req.DomainID); err != nil {
		// Логируем ошибку, но не прерываем создание маршрута
		fmt.Printf("Warning: failed to create nginx update action: %v\n", err)
	}

	return route, nil
}

// GetDomainRoutes получает маршруты домена
func (s *Service) GetDomainRoutes(domainID uuid.UUID) ([]models.DomainRoute, error) {
	rows, err := s.db.Query(`
		SELECT id, domain_id, container_name, port, path, is_active, created, updated
		FROM domain_routes
		WHERE domain_id = $1
		ORDER BY path
	`, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain routes: %v", err)
	}
	defer rows.Close()

	var routes []models.DomainRoute
	for rows.Next() {
		var route models.DomainRoute
		err := rows.Scan(
			&route.ID, &route.DomainID, &route.ContainerName, &route.Port, &route.Path, &route.IsActive, &route.Created, &route.Updated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain route: %v", err)
		}
		routes = append(routes, route)
	}

	return routes, nil
}

// UpdateDomainRoute обновляет маршрут домена
func (s *Service) UpdateDomainRoute(id uuid.UUID, req *models.UpdateDomainRouteRequest) (*models.DomainRoute, error) {
	// Получаем текущий маршрут
	var route models.DomainRoute
	err := s.db.QueryRow(`
		SELECT id, domain_id, container_name, port, path, is_active, created, updated
		FROM domain_routes WHERE id = $1
	`, id).Scan(
		&route.ID, &route.DomainID, &route.ContainerName, &route.Port, &route.Path, &route.IsActive, &route.Created, &route.Updated,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain route: %v", err)
	}

	// Обновляем поля
	if req.ContainerName != nil {
		route.ContainerName = *req.ContainerName
	}
	if req.Port != nil {
		route.Port = *req.Port
	}
	if req.Path != nil {
		route.Path = *req.Path
	}
	if req.IsActive != nil {
		route.IsActive = *req.IsActive
	}

	route.Updated = time.Now()

	_, err = s.db.Exec(`
		UPDATE domain_routes 
		SET container_name = $1, port = $2, path = $3, is_active = $4, updated = $5
		WHERE id = $6
	`, route.ContainerName, route.Port, route.Path, route.IsActive, route.Updated, id)

	if err != nil {
		return nil, fmt.Errorf("failed to update domain route: %v", err)
	}

	// Создаем действие для обновления nginx конфигурации
	if err := s.createNginxUpdateAction(route.DomainID); err != nil {
		// Логируем ошибку, но не прерываем обновление маршрута
		fmt.Printf("Warning: failed to create nginx update action: %v\n", err)
	}

	return &route, nil
}

// DeleteDomainRoute удаляет маршрут домена
func (s *Service) DeleteDomainRoute(id uuid.UUID) error {
	// Получаем domain_id перед удалением
	var domainID uuid.UUID
	err := s.db.QueryRow("SELECT domain_id FROM domain_routes WHERE id = $1", id).Scan(&domainID)
	if err != nil {
		return fmt.Errorf("failed to get domain_id for route: %v", err)
	}

	// Удаляем маршрут
	_, err = s.db.Exec("DELETE FROM domain_routes WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete domain route: %v", err)
	}

	// Создаем действие для обновления nginx конфигурации
	if err := s.createNginxUpdateAction(domainID); err != nil {
		// Логируем ошибку, но не прерываем удаление маршрута
		fmt.Printf("Warning: failed to create nginx update action: %v\n", err)
	}

	return nil
}

// createNginxUpdateAction создает действие для обновления nginx конфигурации
func (s *Service) createNginxUpdateAction(domainID uuid.UUID) error {
	// Получаем информацию о домене
	var domainName string
	var agentID uuid.UUID
	var sslEnabled bool

	err := s.db.QueryRow(`
		SELECT d.name, d.agent_id, d.ssl_enabled
		FROM domains d
		WHERE d.id = $1
	`, domainID).Scan(&domainName, &agentID, &sslEnabled)

	if err != nil {
		return fmt.Errorf("failed to get domain info: %v", err)
	}

	// Получаем все маршруты домена
	rows, err := s.db.Query(`
		SELECT container_name, port, path
		FROM domain_routes
		WHERE domain_id = $1 AND is_active = true
		ORDER BY path
	`, domainID)

	if err != nil {
		return fmt.Errorf("failed to get domain routes: %v", err)
	}
	defer rows.Close()

	var routes []map[string]interface{}
	for rows.Next() {
		var containerName, port, path string
		err := rows.Scan(&containerName, &port, &path)
		if err != nil {
			return fmt.Errorf("failed to scan route: %v", err)
		}

		routes = append(routes, map[string]interface{}{
			"container_name": containerName,
			"port":           port,
			"path":           path,
		})
	}

	// Создаем payload для действия
	payload := map[string]interface{}{
		"domain":      domainName,
		"ssl_enabled": sslEnabled,
		"routes":      routes,
	}

	// Создаем действие
	return s.createAction(agentID, "update_nginx_config", payload)
}

// GetAgentNginxConfig получает конфигурацию nginx для агента
func (s *Service) GetAgentNginxConfig(agentID uuid.UUID) (*models.AgentNginxConfig, error) {
	rows, err := s.db.Query(`
		SELECT d.id, d.name, d.agent_ip, d.ssl_enabled,
		       dr.container_name, dr.port, dr.path, dr.is_active
		FROM domains d
		LEFT JOIN domain_routes dr ON d.id = dr.domain_id
		WHERE d.agent_id = $1 AND d.is_active = true
		ORDER BY d.name, dr.path
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent nginx config: %v", err)
	}
	defer rows.Close()

	config := &models.AgentNginxConfig{
		AgentID: agentID,
		Domains: []models.NginxConfig{},
	}

	domainMap := make(map[string]*models.NginxConfig)

	for rows.Next() {
		var domainID uuid.UUID
		var domainName, agentIP, containerName, port, path string
		var sslEnabled, routeActive bool

		err := rows.Scan(&domainID, &domainName, &agentIP, &sslEnabled, &containerName, &port, &path, &routeActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan nginx config: %v", err)
		}

		// Создаем или получаем конфигурацию домена
		nginxConfig, exists := domainMap[domainName]
		if !exists {
			nginxConfig = &models.NginxConfig{
				Domain:     domainName,
				AgentIP:    agentIP,
				SSLEnabled: sslEnabled,
				Routes:     []models.NginxRoute{},
			}
			domainMap[domainName] = nginxConfig
		}

		// Добавляем маршрут, если он активен
		if routeActive && containerName != "" {
			nginxConfig.Routes = append(nginxConfig.Routes, models.NginxRoute{
				Path:          path,
				ContainerName: containerName,
				Port:          port,
			})
		}
	}

	// Преобразуем map в slice
	for _, nginxConfig := range domainMap {
		config.Domains = append(config.Domains, *nginxConfig)
	}

	return config, nil
}

// GetDomainStatus получает статус домена с информацией о контейнерах
func (s *Service) GetDomainStatus(domainID uuid.UUID) (*models.DomainStatus, error) {
	// Получаем информацию о домене
	var status models.DomainStatus
	err := s.db.QueryRow(`
		SELECT d.id, d.name, d.agent_id, a.name as agent_name, d.agent_ip, d.is_active, d.ssl_enabled
		FROM domains d
		JOIN agents a ON d.agent_id = a.id
		WHERE d.id = $1
	`, domainID).Scan(
		&status.DomainID, &status.DomainName, &status.AgentID, &status.AgentName, &status.AgentIP, &status.IsActive, &status.SSLEnabled,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain status: %v", err)
	}

	// Получаем маршруты с информацией о контейнерах
	rows, err := s.db.Query(`
		SELECT dr.id, dr.container_name, dr.port, dr.path, dr.is_active,
		       c.status as container_status
		FROM domain_routes dr
		LEFT JOIN containers c ON dr.container_name = c.name AND c.agent_id = $1
		WHERE dr.domain_id = $2
		ORDER BY dr.path
	`, status.AgentID, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain routes status: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var routeStatus models.RouteStatus
		var containerStatus sql.NullString

		err := rows.Scan(
			&routeStatus.RouteID, &routeStatus.ContainerName, &routeStatus.Port, &routeStatus.Path, &routeStatus.IsActive,
			&containerStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan route status: %v", err)
		}

		if containerStatus.Valid {
			routeStatus.ContainerStatus = containerStatus.String
		} else {
			routeStatus.ContainerStatus = "not_found"
		}

		status.Routes = append(status.Routes, routeStatus)
	}

	return &status, nil
}

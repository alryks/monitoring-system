package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"monitoring-system/core/server/internal/auth"
	"monitoring-system/core/server/internal/models"
	"monitoring-system/core/server/internal/notifications"
)

type Handlers struct {
	db           *sql.DB
	auth         *auth.Service
	notification *notifications.Service
}

func New(db *sql.DB, authService *auth.Service) *Handlers {
	h := &Handlers{
		db:           db,
		auth:         authService,
		notification: notifications.New(),
	}

	// Создаем админа по умолчанию
	h.createDefaultAdmin()

	return h
}

// createDefaultAdmin создает админа по умолчанию
func (h *Handlers) createDefaultAdmin() {
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		log.Printf("Error checking admin users: %v", err)
		return
	}

	if count == 0 {
		hashedPassword, err := h.auth.HashPassword("admin")
		if err != nil {
			log.Printf("Error hashing admin password: %v", err)
			return
		}

		_, err = h.db.Exec(`
			INSERT INTO users (username, password_hash, role, is_active) 
			VALUES ($1, $2, 'admin', true)
		`, "admin", hashedPassword)

		if err != nil {
			log.Printf("Error creating admin user: %v", err)
		} else {
			log.Println("Default admin user created: admin/admin")
		}
	}
}

// Login обрабатывает авторизацию
// @Summary Аутентификация пользователя
// @Description Выполняет вход в систему и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Данные для входа"
// @Success 200 {object} models.LoginResponse "Успешная аутентификация"
// @Failure 400 {string} string "Неверный JSON"
// @Failure 401 {string} string "Неверные учетные данные"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /login [post]
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var user models.User
	err := h.db.QueryRow(`
		SELECT id, username, password_hash, email, is_active, role, created, last_login
		FROM users 
		WHERE username = $1 AND is_active = true
	`, req.Username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email,
		&user.IsActive, &user.Role, &user.Created, &user.LastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if !h.auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Обновляем время последнего входа
	now := time.Now()
	_, err = h.db.Exec("UPDATE users SET last_login = $1 WHERE id = $2", now, user.ID)
	if err != nil {
		log.Printf("Error updating last login: %v", err)
	}
	user.LastLogin = &now

	token, err := h.auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AgentPing обрабатывает пинги от агентов
// @Summary Пинг от агента
// @Description Получает данные мониторинга от агента и сохраняет их в базе данных, возвращает список невыполненных действий
// @Tags agent-data
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен агента"
// @Param request body models.AgentData true "Данные мониторинга от агента"
// @Success 200 {object} []models.Action "Список невыполненных действий"
// @Failure 400 {string} string "Неверные данные"
// @Failure 401 {string} string "Неверный токен агента"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /agent/ping [post]
func (h *Handlers) AgentPing(w http.ResponseWriter, r *http.Request) {
	// Проверяем Bearer токен агента
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := bearerToken[1]

	// Проверяем существование агента с таким токеном
	var agentID uuid.UUID
	var agentName string
	err := h.db.QueryRow("SELECT id, name FROM agents WHERE token = $1 AND is_active = true", token).Scan(&agentID, &agentName)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid agent token", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Парсим данные от агента
	var agentData models.AgentData
	if err := json.NewDecoder(r.Body).Decode(&agentData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Сохраняем данные в БД
	if err := h.saveAgentData(agentID, &agentData); err != nil {
		log.Printf("Error saving agent data: %v", err)
		http.Error(w, "Error saving data", http.StatusInternalServerError)
		return
	}

	// Проверяем уведомления
	h.checkNotifications(agentID, agentName, &agentData)

	// Получаем список невыполненных действий для агента
	pendingActions, err := h.getPendingActions(agentID)
	if err != nil {
		log.Printf("Error getting pending actions: %v", err)
		http.Error(w, "Error getting actions", http.StatusInternalServerError)
		return
	}

	// Возвращаем список действий
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pendingActions)
}

// saveAgentData сохраняет данные от агента в БД
func (h *Handlers) saveAgentData(agentID uuid.UUID, data *models.AgentData) error {
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновляем время последнего пинга агента
	_, err = tx.Exec("UPDATE agents SET last_ping = now() WHERE id = $1", agentID)
	if err != nil {
		return err
	}

	// Создаем запись пинга
	var pingID uuid.UUID
	err = tx.QueryRow(`
		INSERT INTO agent_pings (agent_id, created) 
		VALUES ($1, now()) 
		RETURNING id
	`, agentID).Scan(&pingID)
	if err != nil {
		return err
	}

	// Сохраняем метрики CPU
	for _, cpu := range data.Metrics.CPU {
		_, err = tx.Exec(`
			INSERT INTO cpu_metrics (ping_id, cpu_name, usage_percent)
			VALUES ($1, $2, $3)
		`, pingID, cpu.Name, cpu.Usage)
		if err != nil {
			return err
		}
	}

	// Сохраняем метрики памяти
	_, err = tx.Exec(`
		INSERT INTO memory_metrics (ping_id, ram_total_mb, ram_usage_mb, swap_total_mb, swap_usage_mb)
		VALUES ($1, $2, $3, $4, $5)
	`, pingID, data.Metrics.Memory.RAM.Total, data.Metrics.Memory.RAM.Usage,
		data.Metrics.Memory.Swap.Total, data.Metrics.Memory.Swap.Usage)
	if err != nil {
		return err
	}

	// Сохраняем метрики дисков
	for _, disk := range data.Metrics.Disk {
		_, err = tx.Exec(`
			INSERT INTO disk_metrics (ping_id, disk_name, read_bytes, write_bytes, reads, writes)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, pingID, disk.Name, disk.ReadBytes, disk.WriteBytes, disk.Reads, disk.Writes)
		if err != nil {
			return err
		}
	}

	// Сохраняем метрики сети
	_, err = tx.Exec(`
		INSERT INTO network_metrics (ping_id, public_ip, sent_bytes, received_bytes)
		VALUES ($1, $2, $3, $4)
	`, pingID, data.Metrics.Network.PublicIP, data.Metrics.Network.Sent, data.Metrics.Network.Received)
	if err != nil {
		return err
	}

	// Сохраняем контейнеры
	for _, container := range data.Docker.Containers {
		containerCreatedAt, _ := time.Parse(time.RFC3339Nano, container.Created)

		var memory *int64
		if container.Memory != nil {
			memoryMB := int64(*container.Memory) // Already in MB from agent
			memory = &memoryMB
		}

		var containerDBID uuid.UUID
		err = tx.QueryRow(`
			INSERT INTO containers (
				ping_id, container_id, name, image_id, status, restart_count, 
				created_at, ip_address, mac_address, cpu_usage_percent, 
				memory_usage_mb, network_sent_bytes, network_received_bytes
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id
		`, pingID, container.ID, container.Name, container.Image, container.Status,
			container.RestartCount, containerCreatedAt, container.IP, container.MAC,
			container.CPU, memory, container.Network.Sent, container.Network.Received).Scan(&containerDBID)
		if err != nil {
			return err
		}

		// Сохраняем логи контейнера
		for _, logLine := range container.Logs {
			// Очищаем строку от null-байтов
			cleanLogLine := strings.ReplaceAll(logLine, "\x00", "")
			if cleanLogLine != "" {
				_, err = tx.Exec(`
				INSERT INTO container_logs (container_id, log_line, timestamp)
				VALUES ($1, $2, now())
			`, containerDBID, cleanLogLine)
				if err != nil {
					return err
				}
			}
		}
	}

	// Сохраняем образы
	for _, image := range data.Docker.Images {
		imageCreatedAt, _ := time.Parse(time.RFC3339, image.Created)

		var imageDBID uuid.UUID
		err = tx.QueryRow(`
		INSERT INTO images (ping_id, image_id, created, size, architecture)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, pingID, image.ID, imageCreatedAt, image.Size, image.Architecture).Scan(&imageDBID)
		if err != nil {
			return err
		}

		// Сохраняем теги образа
		for _, tag := range image.Tags {
			_, err = tx.Exec(`
			INSERT INTO image_tags (image_id, tag)
			VALUES ($1, $2)
		`, imageDBID, tag)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetAgents возвращает список агентов
// @Summary Получить список агентов
// @Description Возвращает список всех активных агентов
// @Tags agents
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Agent "Список агентов"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /agents [get]
func (h *Handlers) GetAgents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT a.id, a.name, a.token, a.is_active, a.created,
			   ap.created as last_ping,
			   COALESCE(nm.public_ip::text, '0.0.0.0') as public_ip
		FROM agents a
		LEFT JOIN (
			SELECT DISTINCT ON (agent_id) agent_id, created, id
			FROM agent_pings
			ORDER BY agent_id, created DESC
		) ap ON a.id = ap.agent_id
		LEFT JOIN network_metrics nm ON ap.id = nm.ping_id
		WHERE a.is_active = true
		ORDER BY a.created DESC
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Инициализируем пустой массив для предотвращения null в JSON
	agents := []models.Agent{}

	for rows.Next() {
		var agent models.Agent
		var publicIP string
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Token, &agent.IsActive, &agent.Created,
			&agent.LastPing, &publicIP,
		)
		if err != nil {
			log.Printf("Error scanning agent: %v", err)
			continue
		}

		if publicIP != "" && publicIP != "0.0.0.0" {
			agent.PublicIP = &publicIP
		}

		// Определяем статус агента
		if agent.LastPing != nil {
			if time.Since(*agent.LastPing) < 2*time.Minute {
				agent.Status = "online"
			} else {
				agent.Status = "offline"
			}
		} else {
			agent.Status = "unknown"
		}

		agents = append(agents, agent)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// CreateAgent создает нового агента
// @Summary Создать нового агента
// @Description Создает нового агента мониторинга и возвращает токен для подключения
// @Tags agents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateAgentRequest true "Данные для создания агента"
// @Success 201 {object} models.Agent "Созданный агент"
// @Failure 400 {string} string "Неверные данные"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /agents [post]
func (h *Handlers) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Agent name is required", http.StatusBadRequest)
		return
	}

	// Генерируем токен для агента
	token, err := h.auth.GenerateAgentToken()
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Создаем агента
	var agent models.Agent
	err = h.db.QueryRow(`
		INSERT INTO agents (name, token, is_active, created)
		VALUES ($1, $2, true, now())
		RETURNING id, name, token, is_active, created
	`, req.Name, token).Scan(
		&agent.ID, &agent.Name, &agent.Token, &agent.IsActive, &agent.Created,
	)
	if err != nil {
		http.Error(w, "Error creating agent", http.StatusInternalServerError)
		return
	}

	agent.Status = "unknown"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// UpdateAgent обновляет агента
// @Summary Обновить агента
// @Description Обновляет имя и статус активности агента
// @Tags agents
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID агента"
// @Param request body object true "Поля для обновления"
// @Success 200 {string} string "Агент успешно обновлен"
// @Failure 400 {string} string "Неверные данные"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /agents/{id} [put]
func (h *Handlers) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     *string `json:"name"`
		IsActive *bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Строим запрос динамически
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *req.Name)
		argCount++
	}

	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *req.IsActive)
		argCount++
	}

	if len(setParts) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("UPDATE agents SET %s WHERE id = $%d", strings.Join(setParts, ", "), argCount)
	args = append(args, agentID)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		http.Error(w, "Error updating agent", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteAgent удаляет агента
// @Summary Удалить агента
// @Description Удаляет агента из системы мониторинга
// @Tags agents
// @Security BearerAuth
// @Param id path string true "ID агента"
// @Success 200 {string} string "Агент успешно удален"
// @Failure 400 {string} string "Неверный ID"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /agents/{id} [delete]
func (h *Handlers) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("DELETE FROM agents WHERE id = $1", agentID)
	if err != nil {
		http.Error(w, "Error deleting agent", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetAgentMetrics возвращает метрики агента
// @Summary Получить метрики агента
// @Description Возвращает историю метрик агента (CPU, память, сеть)
// @Tags agents
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID агента"
// @Param limit query int false "Лимит записей (по умолчанию 50)"
// @Success 200 {array} object "История метрик агента"
// @Failure 400 {string} string "Неверный ID"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /agents/{id}/metrics [get]
func (h *Handlers) GetAgentMetrics(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Получаем параметры пагинации
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "50"
	}
	limit, _ := strconv.Atoi(limitStr)

	// Последние метрики
	rows, err := h.db.Query(`
		SELECT ap.created, nm.public_ip, mm.ram_total_mb, mm.ram_usage_mb,
			   mm.swap_total_mb, mm.swap_usage_mb
		FROM agent_pings ap
		JOIN memory_metrics mm ON ap.id = mm.ping_id
		JOIN network_metrics nm ON ap.id = nm.ping_id
		WHERE ap.agent_id = $1
		ORDER BY ap.created DESC
		LIMIT $2
	`, agentID, limit)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MetricPoint struct {
		Timestamp   time.Time `json:"timestamp"`
		PublicIP    string    `json:"public_ip"`
		RAMTotalMB  int64     `json:"ram_total_mb"`
		RAMUsageMB  int64     `json:"ram_usage_mb"`
		SwapTotalMB int64     `json:"swap_total_mb"`
		SwapUsageMB int64     `json:"swap_usage_mb"`
	}

	var metrics []MetricPoint
	for rows.Next() {
		var metric MetricPoint
		err := rows.Scan(
			&metric.Timestamp, &metric.PublicIP, &metric.RAMTotalMB,
			&metric.RAMUsageMB, &metric.SwapTotalMB, &metric.SwapUsageMB,
		)
		if err != nil {
			log.Printf("Error scanning metric: %v", err)
			continue
		}
		metrics = append(metrics, metric)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetAgentContainers возвращает контейнеры агента
// @Summary Получить контейнеры агента
// @Description Возвращает список контейнеров на конкретном агенте
// @Tags agents
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID агента"
// @Success 200 {array} models.Container "Список контейнеров агента"
// @Failure 400 {string} string "Неверный ID"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /agents/{id}/containers [get]
func (h *Handlers) GetAgentContainers(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Получаем последние контейнеры
	rows, err := h.db.Query(`
		SELECT c.container_id, c.name, c.image_id, c.status, c.restart_count,
			   c.created_at, c.ip_address, c.mac_address, c.cpu_usage_percent,
			   c.memory_usage_mb, c.network_sent_bytes, c.network_received_bytes
		FROM containers c
		JOIN agent_pings ap ON c.ping_id = ap.id
		WHERE ap.agent_id = $1 AND ap.created = (
			SELECT MAX(created) FROM agent_pings WHERE agent_id = $1
		)
		ORDER BY c.name
	`, agentID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var containers []models.Container
	for rows.Next() {
		var container models.Container
		err := rows.Scan(
			&container.ContainerID, &container.Name, &container.ImageID,
			&container.Status, &container.RestartCount, &container.CreatedAt,
			&container.IPAddress, &container.MACAddress, &container.CPUUsagePercent,
			&container.MemoryUsageMB, &container.NetworkSentBytes, &container.NetworkReceivedBytes,
		)
		if err != nil {
			log.Printf("Error scanning container: %v", err)
			continue
		}
		containers = append(containers, container)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

// GetDashboardData возвращает расширенные данные для дашборда
// GetDashboardData возвращает данные для дашборда
// @Summary Получить данные дашборда
// @Description Возвращает KPI метрики, графики использования ресурсов и топ контейнеры
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.DashboardData "Данные дашборда"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /dashboard [get]
func (h *Handlers) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	var dashboard models.DashboardExtended

	// Инициализируем пустые массивы для предотвращения null в JSON
	dashboard.ResourceUsage = []models.ResourceUsagePoint{}
	dashboard.NetworkActivity = []models.NetworkActivityPoint{}
	dashboard.TopContainersCPU = []models.TopContainer{}
	dashboard.TopContainersMemory = []models.TopContainer{}
	dashboard.AgentsSummary = []models.AgentSummary{}

	// Получаем KPI метрики
	var kpis models.KPIMetrics

	// Сначала получаем статистику агентов
	err := h.db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN ap.created > now() - interval '2 minutes' THEN 1 END) as agents_online,
			COUNT(*) as agents_total
		FROM agents a
		LEFT JOIN (
			SELECT DISTINCT ON (agent_id) agent_id, created
			FROM agent_pings
			ORDER BY agent_id, created DESC
		) ap ON a.id = ap.agent_id
		WHERE a.is_active = true
	`).Scan(&kpis.AgentsOnline, &kpis.AgentsTotal)
	if err != nil {
		log.Printf("Error getting agent stats: %v", err)
	}

	// Получаем статистику контейнеров из последних пингов каждого агента
	err = h.db.QueryRow(`
		SELECT 
			COALESCE(COUNT(DISTINCT c.container_id), 0) as containers_total,
			COALESCE(COUNT(DISTINCT CASE WHEN c.status LIKE 'Up %' THEN c.container_id END), 0) as containers_running,
			COALESCE(COUNT(DISTINCT CASE WHEN c.status NOT LIKE 'Up %' THEN c.container_id END), 0) as containers_stopped
		FROM containers c
		JOIN (
			SELECT DISTINCT ON (agent_id) agent_id, id
			FROM agent_pings
			ORDER BY agent_id, created DESC
		) latest_pings ON c.ping_id = latest_pings.id
	`).Scan(&kpis.ContainersTotal, &kpis.ContainersRunning, &kpis.ContainersStopped)
	if err != nil {
		log.Printf("Error getting container stats: %v", err)
	}

	// Получаем средние значения ресурсов отдельным запросом
	err = h.db.QueryRow(`
		SELECT 
			COALESCE(AVG(cm.usage_percent), 0) as avg_cpu_usage,
			COALESCE(AVG(CASE WHEN mm.ram_total_mb > 0 THEN (mm.ram_usage_mb::float / mm.ram_total_mb::float) * 100 END), 0) as avg_memory_usage
		FROM agent_pings ap
		LEFT JOIN cpu_metrics cm ON ap.id = cm.ping_id
		LEFT JOIN memory_metrics mm ON ap.id = mm.ping_id
		WHERE ap.created > now() - interval '5 minutes'
	`).Scan(&kpis.AvgCPUUsage, &kpis.AvgMemoryUsage)
	if err != nil {
		log.Printf("Error getting resource stats: %v", err)
	}
	dashboard.Kpis = kpis

	// Получаем историю использования ресурсов (последние 20 точек)
	resourceUsage, err := h.getResourceUsageHistory()
	if err != nil {
		log.Printf("Error getting resource usage: %v", err)
	} else {
		dashboard.ResourceUsage = resourceUsage
	}

	// Получаем историю сетевой активности
	networkActivity, err := h.getNetworkActivityHistory()
	if err != nil {
		log.Printf("Error getting network activity: %v", err)
	} else {
		dashboard.NetworkActivity = networkActivity
	}

	// Получаем топ 5 контейнеров по CPU
	topCPU, err := h.getTopContainersCPU()
	if err != nil {
		log.Printf("Error getting top CPU containers: %v", err)
	} else {
		dashboard.TopContainersCPU = topCPU
	}

	// Получаем топ 5 контейнеров по памяти
	topMemory, err := h.getTopContainersMemory()
	if err != nil {
		log.Printf("Error getting top memory containers: %v", err)
	} else {
		dashboard.TopContainersMemory = topMemory
	}

	// Получаем сводку по агентам
	agentsSummary, err := h.getAgentsSummary()
	if err != nil {
		log.Printf("Error getting agents summary: %v", err)
	} else {
		dashboard.AgentsSummary = agentsSummary
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (h *Handlers) getAgentList() ([]models.Agent, error) {
	rows, err := h.db.Query(`
		SELECT a.id, a.name, a.token, a.is_active, a.created,
			   ap.created as last_ping, nm.public_ip::text
		FROM agents a
		LEFT JOIN (
			SELECT DISTINCT ON (agent_id) agent_id, created, id
			FROM agent_pings
			ORDER BY agent_id, created DESC
		) ap ON a.id = ap.agent_id
		LEFT JOIN network_metrics nm ON ap.id = nm.ping_id
		WHERE a.is_active = true
		ORDER BY a.created DESC
		LIMIT 10
	`)
	if err != nil {
		return []models.Agent{}, err
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		var agent models.Agent
		var publicIP sql.NullString
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Token, &agent.IsActive, &agent.Created,
			&agent.LastPing, &publicIP,
		)
		if err != nil {
			continue
		}

		if publicIP.Valid && publicIP.String != "0.0.0.0" {
			agent.PublicIP = &publicIP.String
		}

		// Определяем статус
		if agent.LastPing != nil {
			if time.Since(*agent.LastPing) < 2*time.Minute {
				agent.Status = "online"
			} else {
				agent.Status = "offline"
			}
		} else {
			agent.Status = "unknown"
		}

		agents = append(agents, agent)
	}

	// Возвращаем пустой массив вместо nil
	if agents == nil {
		return []models.Agent{}, nil
	}

	return agents, nil
}

func (h *Handlers) getRecentMetrics() ([]models.RecentMetric, error) {
	rows, err := h.db.Query(`
		SELECT a.id, a.name, ap.created, COALESCE(nm.public_ip::text, '0.0.0.0') as public_ip,
			   COALESCE(AVG(cm.usage_percent), 0) as avg_cpu, 
			   COALESCE(mm.ram_usage_mb, 0) as ram_usage_mb, 
			   COALESCE(mm.ram_total_mb, 0) as ram_total_mb
		FROM agents a
		JOIN agent_pings ap ON a.id = ap.agent_id
		LEFT JOIN cpu_metrics cm ON ap.id = cm.ping_id
		LEFT JOIN memory_metrics mm ON ap.id = mm.ping_id
		LEFT JOIN network_metrics nm ON ap.id = nm.ping_id
		WHERE ap.created > now() - interval '1 hour'
		GROUP BY a.id, a.name, ap.created, nm.public_ip, mm.ram_usage_mb, mm.ram_total_mb
		ORDER BY ap.created DESC
		LIMIT 20
	`)
	if err != nil {
		return []models.RecentMetric{}, err
	}
	defer rows.Close()

	var metrics []models.RecentMetric
	for rows.Next() {
		var metric models.RecentMetric
		var ramUsage, ramTotal int64
		err := rows.Scan(
			&metric.AgentID, &metric.AgentName, &metric.Timestamp,
			&metric.PublicIP, &metric.CPUUsage, &ramUsage, &ramTotal,
		)
		if err != nil {
			continue
		}

		if ramTotal > 0 {
			metric.RAMUsage = float64(ramUsage) / float64(ramTotal) * 100
		}

		metrics = append(metrics, metric)
	}

	// Возвращаем пустой массив вместо nil
	if metrics == nil {
		return []models.RecentMetric{}, nil
	}

	return metrics, nil
}

func (h *Handlers) getSystemOverview() (models.SystemOverview, error) {
	var overview models.SystemOverview
	var totalRAM sql.NullInt64

	// Подсчитываем общую статистику из последних метрик
	err := h.db.QueryRow(`
		SELECT COALESCE(COUNT(DISTINCT cm.cpu_name), 0) as total_cpus,
			   COALESCE(SUM(mm.ram_total_mb), 0) as total_ram,
			   COALESCE(COUNT(DISTINCT c.container_id), 0) as total_containers,
			   COALESCE(COUNT(CASE WHEN c.status = 'running' THEN 1 END), 0) as running_containers
		FROM agent_pings ap
		LEFT JOIN cpu_metrics cm ON ap.id = cm.ping_id
		LEFT JOIN memory_metrics mm ON ap.id = mm.ping_id
		LEFT JOIN containers c ON ap.id = c.ping_id
		WHERE ap.created > now() - interval '5 minutes'
	`).Scan(&overview.TotalCPUCores, &totalRAM,
		&overview.TotalContainers, &overview.RunningContainers)

	if err != nil {
		return overview, err
	}

	if totalRAM.Valid {
		overview.TotalRAMMB = totalRAM.Int64
	} else {
		overview.TotalRAMMB = 0
	}

	return overview, err
}

// GetAgentDetail возвращает детальную информацию об агенте
// GetAgentDetail возвращает детальную информацию об агенте
// @Summary Получить детальную информацию об агенте
// @Description Возвращает полную информацию об агенте включая метрики, контейнеры, образы и т.д.
// @Tags agents
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID агента"
// @Success 200 {object} models.AgentDetail "Детальная информация об агенте"
// @Failure 400 {string} string "Неверный ID"
// @Failure 401 {string} string "Не авторизован"
// @Failure 404 {string} string "Агент не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /agents/{id} [get]
func (h *Handlers) GetAgentDetail(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	agentID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Получаем базовую информацию об агенте
	var agent models.Agent
	err = h.db.QueryRow(`
		SELECT a.id, a.name, a.token, a.is_active, a.created,
			   MAX(ap.created) as last_ping,
			   COALESCE(nm.public_ip::text, '0.0.0.0') as public_ip
		FROM agents a
		LEFT JOIN agent_pings ap ON a.id = ap.agent_id
		LEFT JOIN network_metrics nm ON ap.id = nm.ping_id
		WHERE a.id = $1
		GROUP BY a.id, a.name, a.token, a.is_active, a.created, nm.public_ip
	`, agentID).Scan(
		&agent.ID, &agent.Name, &agent.Token, &agent.IsActive, &agent.Created,
		&agent.LastPing, &agent.PublicIP,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Agent not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Определяем статус
	if agent.LastPing != nil {
		if time.Since(*agent.LastPing) < 2*time.Minute {
			agent.Status = "online"
		} else {
			agent.Status = "offline"
		}
	} else {
		agent.Status = "unknown"
	}

	// Получаем текущие метрики
	metrics, err := h.getAgentCurrentMetrics(agentID)
	if err != nil {
		log.Printf("Error getting agent metrics: %v", err)
	}

	// Получаем историю системных метрик
	systemMetrics, err := h.getAgentSystemMetrics(agentID)
	if err != nil {
		log.Printf("Error getting system metrics: %v", err)
	}

	// Получаем контейнеры
	containers, err := h.getAgentContainersDetailed(agentID)
	if err != nil {
		log.Printf("Error getting containers: %v", err)
	}

	// Получаем образы
	images, err := h.getAgentImages(agentID)
	if err != nil {
		log.Printf("Error getting images: %v", err)
	}

	detail := models.AgentDetail{
		Agent:         agent,
		Metrics:       metrics,
		SystemMetrics: systemMetrics,
		Containers:    containers,
		Images:        images,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// GetContainers возвращает все контейнеры со всех агентов
// GetContainers возвращает список контейнеров
// @Summary Получить список контейнеров
// @Description Возвращает список всех контейнеров с фильтрацией
// @Tags containers
// @Produce json
// @Security BearerAuth
// @Param agent_id query string false "ID агента для фильтрации"
// @Param status query string false "Статус для фильтрации"
// @Param search query string false "Поиск по имени или образу"
// @Success 200 {object} models.ContainerListResponse "Список контейнеров"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /containers [get]
func (h *Handlers) GetContainers(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	// Получаем только контейнеры из последнего ping'а для каждого агента (или конкретного агента)
	var args []interface{}
	argCount := 1

	query := `
		WITH latest_pings AS (
			SELECT DISTINCT ON (ap.agent_id) ap.id, ap.agent_id, ap.created
			FROM agent_pings ap
			JOIN agents a ON ap.agent_id = a.id
			WHERE 1=1`

	// Фильтрация по агенту в CTE
	if agentID != "" {
		if agentUUID, err := uuid.Parse(agentID); err == nil {
			query += fmt.Sprintf(" AND a.id = $%d", argCount)
			args = append(args, agentUUID)
			argCount++
		}
	}

	query += `
		ORDER BY ap.agent_id, ap.created DESC
	)
		SELECT c.id, c.ping_id, c.container_id, c.name, c.image_id, c.status, 
			   c.restart_count, c.created_at, c.ip_address, c.mac_address, 
			   c.cpu_usage_percent, c.memory_usage_mb, c.network_sent_bytes, 
			   c.network_received_bytes, 
			   a.id as agent_id, a.name as agent_name, a.is_active, a.created as agent_created, 
			   lp.created as last_ping, '' as public_ip
		FROM containers c
		JOIN latest_pings lp ON c.ping_id = lp.id
		JOIN agents a ON lp.agent_id = a.id
		WHERE 1=1`

	// Дополнительные фильтры для контейнеров
	if status != "" {
		query += fmt.Sprintf(" AND c.status ILIKE $%d", argCount)
		args = append(args, "%"+status+"%")
		argCount++
	}

	if search != "" {
		query += fmt.Sprintf(" AND (c.name ILIKE $%d OR c.container_id ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	query += " ORDER BY c.name"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		log.Printf("Database query error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var containers []models.ContainerDetail
	for rows.Next() {
		var container models.ContainerDetail
		var agentID uuid.UUID
		var agentName string
		var agentIsActive bool
		var agentCreated time.Time
		var agentLastPing time.Time
		var agentPublicIP string

		err := rows.Scan(
			&container.ID, &container.PingID, &container.ContainerID, &container.Name,
			&container.ImageID, &container.Status, &container.RestartCount,
			&container.CreatedAt, &container.IPAddress, &container.MACAddress,
			&container.CPUUsagePercent, &container.MemoryUsageMB,
			&container.NetworkSentBytes, &container.NetworkReceivedBytes,
			&agentID, &agentName, &agentIsActive, &agentCreated, &agentLastPing, &agentPublicIP,
		)
		if err != nil {
			log.Printf("Error scanning container: %v", err)
			continue
		}

		// Определяем статус агента
		status := "unknown"
		if !agentLastPing.IsZero() {
			if time.Since(agentLastPing) < 2*time.Minute {
				status = "online"
			} else {
				status = "offline"
			}
		}

		// Заполняем поля для совместимости с frontend
		container.AgentID = &agentID
		container.AgentName = &agentName

		container.Agent = models.Agent{
			ID:       agentID,
			Name:     agentName,
			IsActive: agentIsActive,
			Created:  agentCreated,
			LastPing: &agentLastPing,
			PublicIP: &agentPublicIP,
			Status:   status,
		}

		containers = append(containers, container)
	}

	response := models.ContainerListResponse{
		Containers: containers,
		Total:      len(containers),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetImages возвращает все образы со всех агентов
// GetImages возвращает список образов
// @Summary Получить список Docker образов
// @Description Возвращает список всех Docker образов с фильтрацией
// @Tags images
// @Produce json
// @Security BearerAuth
// @Param agent_id query string false "ID агента для фильтрации"
// @Param search query string false "Поиск по ID образа или тегу"
// @Success 200 {object} models.ImageListResponse "Список образов"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /images [get]
func (h *Handlers) GetImages(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	search := r.URL.Query().Get("search")

	// Получаем только образы из последнего ping'а для каждого агента (или конкретного агента)
	var args []interface{}
	argCount := 1

	query := `
		WITH latest_pings AS (
			SELECT DISTINCT ON (ap.agent_id) ap.id, ap.agent_id, ap.created
			FROM agent_pings ap
			JOIN agents a ON ap.agent_id = a.id
			WHERE 1=1`

	// Фильтрация по агенту в CTE
	if agentID != "" {
		if agentUUID, err := uuid.Parse(agentID); err == nil {
			query += fmt.Sprintf(" AND a.id = $%d", argCount)
			args = append(args, agentUUID)
			argCount++
		}
	}

	query += `
		ORDER BY ap.agent_id, ap.created DESC
	)
	SELECT i.id, i.ping_id, i.image_id, i.created, i.size, i.architecture,
		   a.id as agent_id, a.name as agent_name,
		   ARRAY_AGG(DISTINCT it.tag) as tags
	FROM images i
	JOIN latest_pings lp ON i.ping_id = lp.id
	JOIN agents a ON lp.agent_id = a.id
	LEFT JOIN image_tags it ON i.id = it.image_id
	WHERE 1=1`

	// Дополнительные фильтры для образов
	if search != "" {
		query += fmt.Sprintf(" AND (i.image_id ILIKE $%d OR it.tag ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
		argCount++
	}

	query += " GROUP BY i.id, i.ping_id, i.image_id, i.created, i.size, i.architecture, a.id, a.name ORDER BY i.created DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var images []models.ImageDetail
	for rows.Next() {
		var image models.ImageDetail
		var agentID uuid.UUID
		var agentName string
		var tags pq.StringArray

		err := rows.Scan(
			&image.ID, &image.PingID, &image.ImageID, &image.Created,
			&image.Size, &image.Architecture, &agentID, &agentName, &tags,
		)
		if err != nil {
			log.Printf("Error scanning image: %v", err)
			continue
		}

		image.Agent = models.Agent{
			ID:   agentID,
			Name: agentName,
		}

		// Преобразуем pq.StringArray в []string и фильтруем NULL значения
		image.Tags = []string{}
		for _, tag := range tags {
			if tag != "" && tag != "NULL" {
				image.Tags = append(image.Tags, tag)
			}
		}

		images = append(images, image)
	}

	response := models.ImageListResponse{
		Images: images,
		Total:  len(images),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetContainerDetail возвращает детальную информацию о контейнере
// GetContainerDetail возвращает детальную информацию о контейнере
// @Summary Получить детальную информацию о контейнере
// @Description Возвращает полную информацию о контейнере включая метрики и историю
// @Tags containers
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID контейнера"
// @Success 200 {object} models.ContainerDetail "Детальная информация о контейнере"
// @Failure 400 {string} string "Неверный ID"
// @Failure 401 {string} string "Не авторизован"
// @Failure 404 {string} string "Контейнер не найден"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /containers/{id} [get]
func (h *Handlers) GetContainerDetail(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	containerID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "Invalid container ID", http.StatusBadRequest)
		return
	}

	// Получаем информацию о контейнере
	var container models.ContainerDetail
	var agentID uuid.UUID
	var agentName string

	err = h.db.QueryRow(`
		SELECT c.id, c.ping_id, c.container_id, c.name, c.image_id, c.status, 
			   c.restart_count, c.created_at, c.ip_address, c.mac_address, 
			   c.cpu_usage_percent, c.memory_usage_mb, c.network_sent_bytes, 
			   c.network_received_bytes, a.id as agent_id, a.name as agent_name
		FROM containers c
		JOIN agent_pings ap ON c.ping_id = ap.id
		JOIN agents a ON ap.agent_id = a.id
		WHERE c.id = $1
	`, containerID).Scan(
		&container.ID, &container.PingID, &container.ContainerID, &container.Name,
		&container.ImageID, &container.Status, &container.RestartCount,
		&container.CreatedAt, &container.IPAddress, &container.MACAddress,
		&container.CPUUsagePercent, &container.MemoryUsageMB,
		&container.NetworkSentBytes, &container.NetworkReceivedBytes,
		&agentID, &agentName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Container not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	container.Agent = models.Agent{
		ID:   agentID,
		Name: agentName,
	}

	// Получаем логи контейнера
	logs, err := h.getContainerLogs(containerID)
	if err != nil {
		log.Printf("Error getting container logs: %v", err)
	}
	container.Logs = logs

	// Получаем историю метрик контейнера
	history, err := h.getContainerHistory(containerID)
	if err != nil {
		log.Printf("Error getting container history: %v", err)
	}
	container.History = history

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(container)
}

// GetContainerLogs возвращает логи контейнера
// @Summary Получить логи контейнера
// @Description Возвращает последние логи контейнера (до 100 записей)
// @Tags containers
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID контейнера"
// @Success 200 {array} models.ContainerLog "Логи контейнера"
// @Failure 400 {string} string "Неверный ID"
// @Failure 401 {string} string "Не авторизован"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /containers/{id}/logs [get]
func (h *Handlers) GetContainerLogs(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	containerID, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "Invalid container ID", http.StatusBadRequest)
		return
	}

	logs, err := h.getContainerLogs(containerID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// Вспомогательные функции

func (h *Handlers) getAgentCurrentMetrics(agentID uuid.UUID) (models.AgentMetrics, error) {
	var metrics models.AgentMetrics

	// Получаем последние метрики CPU
	cpuRows, err := h.db.Query(`
		SELECT cm.cpu_name, cm.usage_percent
		FROM cpu_metrics cm
		JOIN agent_pings ap ON cm.ping_id = ap.id
		WHERE ap.agent_id = $1 AND ap.created = (
			SELECT MAX(created) FROM agent_pings WHERE agent_id = $1
		)
		ORDER BY cm.cpu_name
	`, agentID)
	if err != nil {
		return metrics, err
	}
	defer cpuRows.Close()

	for cpuRows.Next() {
		var cpu models.CPUMetricCurrent
		err := cpuRows.Scan(&cpu.Name, &cpu.Usage)
		if err != nil {
			continue
		}
		metrics.CPU = append(metrics.CPU, cpu)
	}

	// Получаем метрики памяти
	err = h.db.QueryRow(`
		SELECT mm.ram_total_mb, mm.ram_usage_mb, mm.swap_total_mb, mm.swap_usage_mb
		FROM memory_metrics mm
		JOIN agent_pings ap ON mm.ping_id = ap.id
		WHERE ap.agent_id = $1 AND ap.created = (
			SELECT MAX(created) FROM agent_pings WHERE agent_id = $1
		)
	`, agentID).Scan(&metrics.Memory.RAMTotal, &metrics.Memory.RAMUsage, &metrics.Memory.SwapTotal, &metrics.Memory.SwapUsage)
	if err != nil && err != sql.ErrNoRows {
		return metrics, err
	}

	// Вычисляем проценты
	if metrics.Memory.RAMTotal > 0 {
		metrics.Memory.RAMPercent = float64(metrics.Memory.RAMUsage) / float64(metrics.Memory.RAMTotal) * 100
	}
	if metrics.Memory.SwapTotal > 0 {
		metrics.Memory.SwapPercent = float64(metrics.Memory.SwapUsage) / float64(metrics.Memory.SwapTotal) * 100
	}

	// Получаем метрики дисков
	diskRows, err := h.db.Query(`
		SELECT dm.disk_name, dm.read_bytes, dm.write_bytes
		FROM disk_metrics dm
		JOIN agent_pings ap ON dm.ping_id = ap.id
		WHERE ap.agent_id = $1 AND ap.created = (
			SELECT MAX(created) FROM agent_pings WHERE agent_id = $1
		)
		ORDER BY dm.disk_name
	`, agentID)
	if err != nil {
		return metrics, err
	}
	defer diskRows.Close()

	for diskRows.Next() {
		var disk models.DiskMetricCurrent
		err := diskRows.Scan(&disk.Name, &disk.ReadBytes, &disk.WriteBytes)
		if err != nil {
			continue
		}
		metrics.Disk = append(metrics.Disk, disk)
	}

	// Получаем метрики сети
	err = h.db.QueryRow(`
		SELECT nm.public_ip, nm.sent_bytes, nm.received_bytes
		FROM network_metrics nm
		JOIN agent_pings ap ON nm.ping_id = ap.id
		WHERE ap.agent_id = $1 AND ap.created = (
			SELECT MAX(created) FROM agent_pings WHERE agent_id = $1
		)
	`, agentID).Scan(&metrics.Network.PublicIP, &metrics.Network.SentBytes, &metrics.Network.ReceivedBytes)
	if err != nil && err != sql.ErrNoRows {
		return metrics, err
	}

	return metrics, nil
}

func (h *Handlers) getAgentSystemMetrics(agentID uuid.UUID) ([]models.SystemMetric, error) {
	rows, err := h.db.Query(`
		SELECT ap.created, 
			   COALESCE(AVG(cm.usage_percent), 0) as avg_cpu,
			   COALESCE(AVG(CASE WHEN mm.ram_total_mb > 0 THEN (mm.ram_usage_mb::float / mm.ram_total_mb::float) * 100 END), 0) as avg_ram,
			   COALESCE(SUM(dm.read_bytes), 0) as disk_read,
			   COALESCE(SUM(dm.write_bytes), 0) as disk_write,
			   COALESCE(SUM(c.network_sent_bytes), 0) as network_sent,
			   COALESCE(SUM(c.network_received_bytes), 0) as network_received,
			   COALESCE(nm.public_ip, '0.0.0.0') as public_ip
		FROM agent_pings ap
		LEFT JOIN cpu_metrics cm ON ap.id = cm.ping_id
		LEFT JOIN memory_metrics mm ON ap.id = mm.ping_id
		LEFT JOIN disk_metrics dm ON ap.id = dm.ping_id
		LEFT JOIN containers c ON ap.id = c.ping_id
		LEFT JOIN network_metrics nm ON ap.id = nm.ping_id
		WHERE ap.agent_id = $1 AND ap.created > now() - interval '1 hour'
		GROUP BY ap.created, nm.public_ip
		ORDER BY ap.created DESC
		LIMIT 50
	`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.SystemMetric
	for rows.Next() {
		var metric models.SystemMetric
		err := rows.Scan(&metric.Timestamp, &metric.CPUUsage, &metric.RAMUsage,
			&metric.DiskRead, &metric.DiskWrite, &metric.NetworkSent, &metric.NetworkReceived, &metric.PublicIP)
		if err != nil {
			continue
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (h *Handlers) getAgentContainersDetailed(agentID uuid.UUID) ([]models.ContainerDetail, error) {
	rows, err := h.db.Query(`
		SELECT c.id, c.ping_id, c.container_id, c.name, c.image_id, c.status, 
			   c.restart_count, c.created_at, c.ip_address, c.mac_address, 
			   c.cpu_usage_percent, c.memory_usage_mb, c.network_sent_bytes, 
			   c.network_received_bytes
		FROM containers c
		JOIN agent_pings ap ON c.ping_id = ap.id
		WHERE ap.agent_id = $1 AND ap.created = (
			SELECT MAX(created) FROM agent_pings WHERE agent_id = $1
		)
		ORDER BY c.name
	`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []models.ContainerDetail
	for rows.Next() {
		var container models.ContainerDetail
		err := rows.Scan(
			&container.ID, &container.PingID, &container.ContainerID, &container.Name,
			&container.ImageID, &container.Status, &container.RestartCount,
			&container.CreatedAt, &container.IPAddress, &container.MACAddress,
			&container.CPUUsagePercent, &container.MemoryUsageMB,
			&container.NetworkSentBytes, &container.NetworkReceivedBytes,
		)
		if err != nil {
			continue
		}
		containers = append(containers, container)
	}

	return containers, nil
}

func (h *Handlers) getAgentImages(agentID uuid.UUID) ([]models.ImageDetail, error) {
	rows, err := h.db.Query(`
		SELECT i.id, i.ping_id, i.image_id, i.created, i.size, i.architecture,
			   ARRAY_AGG(DISTINCT it.tag) as tags
		FROM images i
		JOIN agent_pings ap ON i.ping_id = ap.id
		LEFT JOIN image_tags it ON i.id = it.image_id
		WHERE ap.agent_id = $1 AND ap.created = (
			SELECT MAX(created) FROM agent_pings WHERE agent_id = $1
		)
		GROUP BY i.id, i.ping_id, i.image_id, i.created, i.size, i.architecture
		ORDER BY i.created DESC
	`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.ImageDetail
	for rows.Next() {
		var image models.ImageDetail
		var tags []string
		err := rows.Scan(
			&image.ID, &image.PingID, &image.ImageID, &image.Created,
			&image.Size, &image.Architecture, &tags,
		)
		if err != nil {
			continue
		}

		if tags != nil {
			image.Tags = tags
		} else {
			image.Tags = []string{}
		}

		images = append(images, image)
	}

	return images, nil
}

func (h *Handlers) getContainerLogs(containerID uuid.UUID) ([]models.ContainerLog, error) {
	rows, err := h.db.Query(`
		SELECT id, log_line, timestamp FROM container_logs 
		WHERE container_id = $1 
		ORDER BY timestamp DESC 
		LIMIT 100
	`, containerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ContainerLog
	for rows.Next() {
		var log models.ContainerLog
		err := rows.Scan(&log.ID, &log.LogLine, &log.Timestamp)
		if err != nil {
			continue
		}
		log.ContainerID = containerID
		logs = append(logs, log)
	}

	return logs, nil
}

func (h *Handlers) getContainerHistory(containerID uuid.UUID) ([]models.ContainerMetric, error) {
	rows, err := h.db.Query(`
		SELECT ap.created, c.cpu_usage_percent, c.memory_usage_mb
		FROM containers c
		JOIN agent_pings ap ON c.ping_id = ap.id
		WHERE c.container_id = (
			SELECT container_id FROM containers WHERE id = $1
		)
		ORDER BY ap.created DESC
		LIMIT 50
	`, containerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.ContainerMetric
	for rows.Next() {
		var metric models.ContainerMetric
		err := rows.Scan(&metric.Timestamp, &metric.CPUUsage, &metric.MemoryUsage)
		if err != nil {
			continue
		}
		history = append(history, metric)
	}

	return history, nil
}

func (h *Handlers) getTopContainersCPU() ([]models.TopContainer, error) {
	rows, err := h.db.Query(`
		SELECT DISTINCT c.name, a.name as agent_name, c.cpu_usage_percent, 
			   COALESCE(c.memory_usage_mb, 0), c.status
		FROM containers c
		JOIN (
			SELECT DISTINCT ON (agent_id) agent_id, id
			FROM agent_pings
			ORDER BY agent_id, created DESC
		) latest_pings ON c.ping_id = latest_pings.id
		JOIN agents a ON latest_pings.agent_id = a.id
		WHERE c.cpu_usage_percent IS NOT NULL
		ORDER BY c.cpu_usage_percent DESC
		LIMIT 5
	`)
	if err != nil {
		return []models.TopContainer{}, err
	}
	defer rows.Close()

	var containers []models.TopContainer
	for rows.Next() {
		var container models.TopContainer
		err := rows.Scan(
			&container.Name, &container.AgentName, &container.CPUUsage,
			&container.MemoryUsage, &container.Status,
		)
		if err != nil {
			continue
		}

		containers = append(containers, container)
	}

	if containers == nil {
		return []models.TopContainer{}, nil
	}

	return containers, nil
}

func (h *Handlers) getTopContainersRAM() ([]models.TopContainer, error) {
	rows, err := h.db.Query(`
		SELECT 
			c.name,
			a.name as agent_name,
			c.memory_usage_mb,
			c.status
		FROM containers c
		JOIN agent_pings ap ON c.ping_id = ap.id
		JOIN agents a ON ap.agent_id = a.id
		WHERE c.memory_usage_mb IS NOT NULL
		  AND ap.created > now() - interval '5 minutes'
		ORDER BY c.memory_usage_mb DESC
		LIMIT 5
	`)
	if err != nil {
		return []models.TopContainer{}, err
	}
	defer rows.Close()

	var containers []models.TopContainer
	for rows.Next() {
		var container models.TopContainer
		err := rows.Scan(
			&container.Name, &container.AgentName,
			&container.MemoryUsage, &container.Status,
		)
		if err != nil {
			continue
		}
		containers = append(containers, container)
	}

	if containers == nil {
		return []models.TopContainer{}, nil
	}

	return containers, nil
}

// getResourceUsageHistory возвращает историю использования ресурсов
func (h *Handlers) getResourceUsageHistory() ([]models.ResourceUsagePoint, error) {
	rows, err := h.db.Query(`
		SELECT 
			DATE_TRUNC('minute', ap.created) as timestamp,
			COALESCE(AVG(cm.usage_percent), 0) as avg_cpu,
			COALESCE(AVG(CASE WHEN mm.ram_total_mb > 0 THEN (mm.ram_usage_mb::float / mm.ram_total_mb::float) * 100 END), 0) as avg_memory
		FROM agent_pings ap
		LEFT JOIN cpu_metrics cm ON ap.id = cm.ping_id
		LEFT JOIN memory_metrics mm ON ap.id = mm.ping_id
		WHERE ap.created > now() - interval '2 hours'
		GROUP BY DATE_TRUNC('minute', ap.created)
		ORDER BY timestamp DESC
		LIMIT 20
	`)
	if err != nil {
		return []models.ResourceUsagePoint{}, err
	}
	defer rows.Close()

	var points []models.ResourceUsagePoint
	for rows.Next() {
		var point models.ResourceUsagePoint
		err := rows.Scan(&point.Timestamp, &point.CPU, &point.Memory)
		if err != nil {
			continue
		}
		points = append(points, point)
	}

	// Реверсируем массив чтобы показать от старых к новым
	for i := len(points)/2 - 1; i >= 0; i-- {
		opp := len(points) - 1 - i
		points[i], points[opp] = points[opp], points[i]
	}

	if points == nil {
		return []models.ResourceUsagePoint{}, nil
	}

	return points, nil
}

// getNetworkActivityHistory возвращает историю сетевой активности
func (h *Handlers) getNetworkActivityHistory() ([]models.NetworkActivityPoint, error) {
	rows, err := h.db.Query(`
		SELECT 
			DATE_TRUNC('minute', ap.created) as timestamp,
			COALESCE(SUM(c.network_sent_bytes), 0) as total_sent,
			COALESCE(SUM(c.network_received_bytes), 0) as total_received
		FROM agent_pings ap
		LEFT JOIN containers c ON ap.id = c.ping_id
		WHERE ap.created > now() - interval '2 hours'
		GROUP BY DATE_TRUNC('minute', ap.created)
		ORDER BY timestamp DESC
		LIMIT 20
	`)
	if err != nil {
		return []models.NetworkActivityPoint{}, err
	}
	defer rows.Close()

	var points []models.NetworkActivityPoint
	for rows.Next() {
		var point models.NetworkActivityPoint
		err := rows.Scan(&point.Timestamp, &point.Sent, &point.Received)
		if err != nil {
			continue
		}
		points = append(points, point)
	}

	// Реверсируем массив чтобы показать от старых к новым
	for i := len(points)/2 - 1; i >= 0; i-- {
		opp := len(points) - 1 - i
		points[i], points[opp] = points[opp], points[i]
	}

	if points == nil {
		return []models.NetworkActivityPoint{}, nil
	}

	return points, nil
}

// getTopContainersMemory возвращает топ контейнеров по использованию памяти
func (h *Handlers) getTopContainersMemory() ([]models.TopContainer, error) {
	rows, err := h.db.Query(`
		SELECT 
			c.name,
			a.name as agent_name,
			c.memory_usage_mb,
			c.status
		FROM containers c
		JOIN agent_pings ap ON c.ping_id = ap.id
		JOIN agents a ON ap.agent_id = a.id
		WHERE c.memory_usage_mb IS NOT NULL
		  AND ap.created > now() - interval '5 minutes'
		ORDER BY c.memory_usage_mb DESC
		LIMIT 5
	`)
	if err != nil {
		return []models.TopContainer{}, err
	}
	defer rows.Close()

	var containers []models.TopContainer
	for rows.Next() {
		var container models.TopContainer
		err := rows.Scan(
			&container.Name, &container.AgentName,
			&container.MemoryUsage, &container.Status,
		)
		if err != nil {
			continue
		}
		containers = append(containers, container)
	}

	if containers == nil {
		return []models.TopContainer{}, nil
	}

	return containers, nil
}

// getAgentsSummary возвращает сводку по агентам
func (h *Handlers) getAgentsSummary() ([]models.AgentSummary, error) {
	rows, err := h.db.Query(`
		SELECT 
			a.id,
			a.name,
			latest_ping.created as last_ping,
			COALESCE(COUNT(DISTINCT c.container_id), 0) as containers_count,
			COALESCE(AVG(cm.usage_percent), 0) as avg_cpu,
			COALESCE(AVG(CASE WHEN mm.ram_total_mb > 0 THEN (mm.ram_usage_mb::float / mm.ram_total_mb::float) * 100 END), 0) as avg_memory
		FROM agents a
		LEFT JOIN (
			SELECT DISTINCT ON (agent_id) agent_id, created, id
			FROM agent_pings
			ORDER BY agent_id, created DESC
		) latest_ping ON a.id = latest_ping.agent_id
		LEFT JOIN containers c ON latest_ping.id = c.ping_id
		LEFT JOIN cpu_metrics cm ON latest_ping.id = cm.ping_id
		LEFT JOIN memory_metrics mm ON latest_ping.id = mm.ping_id
		WHERE a.is_active = true
		GROUP BY a.id, a.name, latest_ping.created
		ORDER BY a.name
	`)
	if err != nil {
		return []models.AgentSummary{}, err
	}
	defer rows.Close()

	var agents []models.AgentSummary
	for rows.Next() {
		var agent models.AgentSummary
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.LastPing,
			&agent.Containers, &agent.CPUUsage, &agent.MemoryUsage,
		)
		if err != nil {
			continue
		}

		// Определяем статус агента
		if agent.LastPing != nil {
			if time.Since(*agent.LastPing) < 2*time.Minute {
				agent.Status = "online"
			} else {
				agent.Status = "offline"
			}
		} else {
			agent.Status = "unknown"
		}

		agents = append(agents, agent)
	}

	if agents == nil {
		return []models.AgentSummary{}, nil
	}

	return agents, nil
}

// getPendingActions получает список невыполненных действий для агента
func (h *Handlers) getPendingActions(agentID uuid.UUID) ([]models.Action, error) {
	rows, err := h.db.Query(`
		SELECT id, agent_id, type, payload, status, created, completed, response, error
		FROM actions 
		WHERE agent_id = $1 AND status = $2
		ORDER BY created ASC
	`, agentID, models.ActionStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []models.Action
	for rows.Next() {
		var action models.Action
		var payloadJSON []byte
		err := rows.Scan(
			&action.ID, &action.AgentID, &action.Type, &payloadJSON,
			&action.Status, &action.Created, &action.Completed,
			&action.Response, &action.Error,
		)
		if err != nil {
			return nil, err
		}

		// Парсим JSON payload
		if err := json.Unmarshal(payloadJSON, &action.Payload); err != nil {
			return nil, err
		}

		actions = append(actions, action)
	}

	return actions, nil
}

// CreateAction создает новое действие для агента
// @Summary Создание действия
// @Description Создает новое действие для выполнения агентом
// @Tags actions
// @Accept json
// @Produce json
// @Param request body models.CreateActionRequest true "Данные действия"
// @Success 201 {object} models.Action "Действие создано"
// @Failure 400 {string} string "Неверные данные"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /actions [post]
func (h *Handlers) CreateAction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Проверяем существование агента
	agentID, err := uuid.Parse(req.AgentID)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	var agentExists bool
	err = h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1 AND is_active = true)", agentID).Scan(&agentExists)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !agentExists {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	// Создаем действие
	payloadJSON, err := json.Marshal(req.Payload)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var action models.Action
	err = h.db.QueryRow(`
		INSERT INTO actions (agent_id, type, payload, status, created)
		VALUES ($1, $2, $3, $4, now())
		RETURNING id, agent_id, type, payload, status, created, completed, response, error
	`, agentID, req.Type, payloadJSON, models.ActionStatusPending).Scan(
		&action.ID, &action.AgentID, &action.Type, &payloadJSON,
		&action.Status, &action.Created, &action.Completed,
		&action.Response, &action.Error,
	)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Парсим payload обратно
	if err := json.Unmarshal(payloadJSON, &action.Payload); err != nil {
		http.Error(w, "Error processing action", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(action)
}

// GetActions получает список действий
// @Summary Получение списка действий
// @Description Получает список действий с фильтрацией
// @Tags actions
// @Produce json
// @Param agent_id query string false "ID агента"
// @Param status query string false "Статус действия"
// @Param type query string false "Тип действия"
// @Success 200 {object} models.ActionListResponse "Список действий"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /actions [get]
func (h *Handlers) GetActions(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	status := r.URL.Query().Get("status")
	actionType := r.URL.Query().Get("type")

	query := `
		SELECT id, agent_id, type, payload, status, created, completed, response, error
		FROM actions 
		WHERE 1=1
	`
	var args []interface{}
	var conditions []string
	argCount := 1

	if agentID != "" {
		conditions = append(conditions, fmt.Sprintf("agent_id = $%d", argCount))
		args = append(args, agentID)
		argCount++
	}

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
		args = append(args, status)
		argCount++
	}

	if actionType != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argCount))
		args = append(args, actionType)
		argCount++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var actions []models.Action
	for rows.Next() {
		var action models.Action
		var payloadJSON []byte
		err := rows.Scan(
			&action.ID, &action.AgentID, &action.Type, &payloadJSON,
			&action.Status, &action.Created, &action.Completed,
			&action.Response, &action.Error,
		)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Парсим JSON payload
		if err := json.Unmarshal(payloadJSON, &action.Payload); err != nil {
			http.Error(w, "Error processing action", http.StatusInternalServerError)
			return
		}

		actions = append(actions, action)
	}

	response := models.ActionListResponse{
		Actions: actions,
		Total:   len(actions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateActionStatus обновляет статус действия
// @Summary Обновление статуса действия
// @Description Обновляет статус действия и сохраняет ответ от агента
// @Tags actions
// @Accept json
// @Produce json
// @Param id path string true "ID действия"
// @Param Authorization header string true "Bearer токен агента"
// @Param request body models.ActionResponse true "Ответ агента"
// @Success 200 {object} models.Action "Действие обновлено"
// @Failure 400 {string} string "Неверные данные"
// @Failure 401 {string} string "Неверный токен агента"
// @Failure 404 {string} string "Действие не найдено"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /actions/{id}/status [put]
func (h *Handlers) UpdateActionStatus(w http.ResponseWriter, r *http.Request) {
	// Проверяем Bearer токен агента
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := bearerToken[1]

	// Проверяем существование агента с таким токеном
	var agentID uuid.UUID
	err := h.db.QueryRow("SELECT id FROM agents WHERE token = $1 AND is_active = true", token).Scan(&agentID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid agent token", http.StatusUnauthorized)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	actionID := chi.URLParam(r, "id")
	if actionID == "" {
		http.Error(w, "Action ID required", http.StatusBadRequest)
		return
	}

	var req models.ActionResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Проверяем, что действие принадлежит этому агенту
	var actionAgentID uuid.UUID
	err = h.db.QueryRow("SELECT agent_id FROM actions WHERE id = $1", actionID).Scan(&actionAgentID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Action not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if actionAgentID != agentID {
		http.Error(w, "Action does not belong to this agent", http.StatusForbidden)
		return
	}

	// Обновляем статус действия
	var completed *time.Time
	if req.Status == models.ActionStatusCompleted || req.Status == models.ActionStatusFailed {
		now := time.Now()
		completed = &now
	}

	_, err = h.db.Exec(`
		UPDATE actions 
		SET status = $1, completed = $2, response = $3, error = $4
		WHERE id = $5
	`, req.Status, completed, req.Response, req.Error, actionID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetNotificationSettings получает настройки уведомлений
// @Summary Получение настроек уведомлений
// @Description Получает текущие настройки уведомлений
// @Tags notifications
// @Produce json
// @Success 200 {object} models.NotificationSettings "Настройки уведомлений"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /notifications/settings [get]
func (h *Handlers) GetNotificationSettings(w http.ResponseWriter, r *http.Request) {
	settings := h.notification.GetSettings()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateNotificationSettings обновляет настройки уведомлений
// @Summary Обновление настроек уведомлений
// @Description Обновляет настройки уведомлений
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body models.NotificationSettings true "Настройки уведомлений"
// @Success 200 {object} models.NotificationSettings "Настройки обновлены"
// @Failure 400 {string} string "Неверные данные"
// @Failure 500 {string} string "Ошибка сервера"
// @Router /notifications/settings [post]
func (h *Handlers) UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	var settings models.NotificationSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Обновляем настройки в сервисе
	h.notification.UpdateSettings(&settings)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// SendTestNotification отправляет тестовое уведомление
// @Summary Отправка тестового уведомления
// @Description Отправляет тестовое уведомление в Telegram
// @Tags notifications
// @Produce json
// @Success 200 {object} map[string]string "Уведомление отправлено"
// @Failure 400 {string} string "Не настроен токен бота"
// @Failure 500 {string} string "Ошибка отправки"
// @Router /notifications/test [post]
func (h *Handlers) SendTestNotification(w http.ResponseWriter, r *http.Request) {
	err := h.notification.SendTestNotification()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Test notification sent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// checkNotifications проверяет условия для отправки уведомлений
func (h *Handlers) checkNotifications(agentID uuid.UUID, agentName string, agentData *models.AgentData) {
	// Проверяем CPU
	if len(agentData.Metrics.CPU) > 0 {
		totalCPU := 0.0
		for _, cpu := range agentData.Metrics.CPU {
			totalCPU += cpu.Usage
		}
		avgCPU := totalCPU / float64(len(agentData.Metrics.CPU))

		if err := h.notification.CheckCPUThreshold(agentName, avgCPU); err != nil {
			log.Printf("Error sending CPU threshold notification: %v", err)
		}
	}

	// Проверяем RAM
	if agentData.Metrics.Memory.RAM.Total > 0 {
		ramUsagePercent := float64(agentData.Metrics.Memory.RAM.Usage) / float64(agentData.Metrics.Memory.RAM.Total) * 100

		if err := h.notification.CheckRAMThreshold(agentName, ramUsagePercent); err != nil {
			log.Printf("Error sending RAM threshold notification: %v", err)
		}
	}

	// Проверяем контейнеры
	for _, container := range agentData.Docker.Containers {
		if container.Status == "exited" || container.Status == "stopped" {
			if err := h.notification.CheckContainerStopped(container.Name, agentName); err != nil {
				log.Printf("Error sending container stopped notification: %v", err)
			}
		}
	}
}

// CheckOfflineAgents проверяет агентов, которые не отвечают, и отправляет уведомления
func (h *Handlers) CheckOfflineAgents() {
	// Находим агентов, которые не отвечали более 60 секунд
	rows, err := h.db.Query(`
		SELECT id, name 
		FROM agents 
		WHERE is_active = true 
		AND (last_ping IS NULL OR last_ping < NOW() - INTERVAL '60 seconds')
	`)
	if err != nil {
		log.Printf("Error checking offline agents: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var agentID uuid.UUID
		var agentName string
		if err := rows.Scan(&agentID, &agentName); err != nil {
			log.Printf("Error scanning agent: %v", err)
			continue
		}

		// Отправляем уведомление о недоступности агента
		if err := h.notification.CheckAgentOffline(agentName); err != nil {
			log.Printf("Error sending agent offline notification: %v", err)
		}
	}
}

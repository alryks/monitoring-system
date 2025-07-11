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

	"monitoring-system/core/server/internal/auth"
	"monitoring-system/core/server/internal/models"
)

type Handlers struct {
	db   *sql.DB
	auth *auth.Service
}

func New(db *sql.DB, authService *auth.Service) *Handlers {
	h := &Handlers{
		db:   db,
		auth: authService,
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
	err := h.db.QueryRow("SELECT id FROM agents WHERE token = $1 AND is_active = true", token).Scan(&agentID)
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

	w.WriteHeader(http.StatusOK)
}

// saveAgentData сохраняет данные от агента в БД
func (h *Handlers) saveAgentData(agentID uuid.UUID, data *models.AgentData) error {
	tx, err := h.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
			memoryMB := int64(*container.Memory / 1024 / 1024) // Convert to MB
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

		// Сохраняем сети контейнера
		for _, network := range container.Network.Networks {
			_, err = tx.Exec(`
				INSERT INTO container_networks (container_id, network_name)
				VALUES ($1, $2)
			`, containerDBID, network)
			if err != nil {
				return err
			}
		}

		// Сохраняем тома контейнера
		for _, volume := range container.Volumes {
			_, err = tx.Exec(`
				INSERT INTO container_volumes (container_id, volume_name)
				VALUES ($1, $2)
			`, containerDBID, volume)
			if err != nil {
				return err
			}
		}

		// Сохраняем логи контейнера
		for i, logLine := range container.Logs {
			_, err = tx.Exec(`
				INSERT INTO container_logs (container_id, log_line, line_number)
				VALUES ($1, $2, $3)
			`, containerDBID, logLine, i+1)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetAgents возвращает список агентов
func (h *Handlers) GetAgents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT a.id, a.name, a.token, a.is_active, a.created,
			   MAX(ap.created) as last_ping,
			   COALESCE(nm.public_ip::text, '0.0.0.0') as public_ip
		FROM agents a
		LEFT JOIN agent_pings ap ON a.id = ap.agent_id
		LEFT JOIN network_metrics nm ON ap.id = nm.ping_id
		GROUP BY a.id, a.name, a.token, a.is_active, a.created, nm.public_ip
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

// GetDashboardData возвращает данные для дашборда
func (h *Handlers) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	var dashboard models.DashboardData

	// Инициализируем пустые массивы для предотвращения null в JSON
	dashboard.Agents = []models.Agent{}
	dashboard.RecentMetrics = []models.RecentMetric{}

	// Получаем статистику агентов
	err := h.db.QueryRow(`
		SELECT COUNT(*) as total,
			   COUNT(CASE WHEN ap.created > now() - interval '2 minutes' THEN 1 END) as online,
			   COUNT(CASE WHEN ap.created <= now() - interval '2 minutes' OR ap.created IS NULL THEN 1 END) as offline
		FROM agents a
		LEFT JOIN (
			SELECT DISTINCT ON (agent_id) agent_id, created
			FROM agent_pings
			ORDER BY agent_id, created DESC
		) ap ON a.id = ap.agent_id
		WHERE a.is_active = true
	`).Scan(&dashboard.TotalAgents, &dashboard.OnlineAgents, &dashboard.OfflineAgents)
	if err != nil {
		log.Printf("Error getting agent stats: %v", err)
	}

	// Получаем список агентов
	agents, err := h.getAgentList()
	if err != nil {
		log.Printf("Error getting agent list: %v", err)
	} else {
		dashboard.Agents = agents
	}

	// Получаем последние метрики
	recentMetrics, err := h.getRecentMetrics()
	if err != nil {
		log.Printf("Error getting recent metrics: %v", err)
	} else {
		dashboard.RecentMetrics = recentMetrics
	}

	// Получаем общую статистику системы
	systemOverview, err := h.getSystemOverview()
	if err != nil {
		log.Printf("Error getting system overview: %v", err)
	} else {
		dashboard.SystemOverview = systemOverview
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

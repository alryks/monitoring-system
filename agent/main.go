package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"bufio"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pion/stun"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	gopsutilnet "github.com/shirou/gopsutil/v3/net"
)

// Action представляет действие от сервера
type Action struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Status    string                 `json:"status"`
	Created   string                 `json:"created"`
	Completed *string                `json:"completed"`
	Response  *string                `json:"response"`
	Error     *string                `json:"error"`
}

// ActionResponse представляет ответ агента на действие
type ActionResponse struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Response *string `json:"response"`
	Error    *string `json:"error"`
}

// Константы для типов действий
const (
	ActionTypeStartContainer  = "start_container"
	ActionTypeStopContainer   = "stop_container"
	ActionTypeRemoveContainer = "remove_container"
	ActionTypeRemoveImage     = "remove_image"
	ActionTypeRestartNginx    = "restart_nginx"
	ActionTypeWriteFile       = "write_file"
)

// Константы для статусов действий
const (
	ActionStatusPending   = "pending"
	ActionStatusCompleted = "completed"
	ActionStatusFailed    = "failed"
)

type AgentData struct {
	Metrics Metrics    `json:"metrics"`
	Docker  DockerInfo `json:"docker"`
}

type Metrics struct {
	CPU     []CPUInfo   `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Disk    []DiskInfo  `json:"disk"`
	Network NetworkInfo `json:"network"`
}

type CPUInfo struct {
	Name  string  `json:"name"`
	Usage float64 `json:"usage"`
}

type MemoryInfo struct {
	RAM  RAMInfo  `json:"ram"`
	Swap SwapInfo `json:"swap"`
}

type RAMInfo struct {
	Total uint64 `json:"total"`
	Usage uint64 `json:"usage"`
}

type SwapInfo struct {
	Total uint64 `json:"total"`
	Usage uint64 `json:"usage"`
}

type DiskInfo struct {
	Name       string `json:"name"`
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	Reads      uint64 `json:"reads"`
	Writes     uint64 `json:"writes"`
}

type NetworkInfo struct {
	PublicIP string `json:"public_ip"`
	Sent     uint64 `json:"sent"`
	Received uint64 `json:"received"`
}

type DockerInfo struct {
	Containers []ContainerInfo `json:"containers"`
	Images     []ImageInfo     `json:"images"`
}

type ContainerInfo struct {
	ID           string               `json:"id"`
	Created      string               `json:"created"`
	Status       string               `json:"status"`
	RestartCount int                  `json:"restart_count"`
	Image        string               `json:"image"`
	Name         string               `json:"name"`
	IP           *string              `json:"ip"`
	MAC          *string              `json:"mac"`
	CPU          *float64             `json:"cpu"`
	Memory       *uint64              `json:"memory"`
	Network      ContainerNetworkInfo `json:"network"`
	Logs         []string             `json:"logs"`
}

type ContainerNetworkInfo struct {
	Sent     *uint64  `json:"sent"`
	Received *uint64  `json:"received"`
	Networks []string `json:"networks"`
}

type ImageInfo struct {
	ID      string   `json:"id"`
	Created string   `json:"created"`
	Size    int64    `json:"size"`
	Tags    []string `json:"tags"`
}

func main() {
	// Читаем переменные окружения
	url := os.Getenv("URL")
	if url == "" {
		log.Fatal("URL environment variable is required")
	}

	token := os.Getenv("TOKEN")
	if token == "" {
		log.Fatal("TOKEN environment variable is required")
	}

	intervalStr := os.Getenv("INTERVAL")
	if intervalStr == "" {
		intervalStr = "5"
	}

	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Fatal("Invalid INTERVAL value:", err)
	}

	// Создаем Docker клиент
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal("Failed to create Docker client:", err)
	}
	defer dockerClient.Close()

	log.Printf("Agent started. Sending data to %s every %d seconds", url, interval)

	// Основной цикл
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		data, err := collectData(dockerClient)
		if err != nil {
			log.Printf("Error collecting data: %v", err)
		} else {
			actions, err := sendData(url, token, data)
			if err != nil {
				log.Printf("Error sending data: %v", err)
			} else {
				log.Println("Data sent successfully")

				// Обрабатываем полученные действия
				if len(actions) > 0 {
					log.Printf("Received %d actions to process", len(actions))
					for _, action := range actions {
						if err := processAction(dockerClient, action); err != nil {
							log.Printf("Error processing action %s: %v", action.ID, err)
						}
					}
				}
			}
		}

		<-ticker.C
	}
}

func collectData(dockerClient *client.Client) (*AgentData, error) {
	ctx := context.Background()

	// Собираем системные метрики
	metrics, err := collectSystemMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to collect system metrics: %v", err)
	}

	// Собираем Docker метрики
	dockerInfo, err := collectDockerMetrics(ctx, dockerClient)
	if err != nil {
		return nil, fmt.Errorf("failed to collect docker metrics: %v", err)
	}

	return &AgentData{
		Metrics: *metrics,
		Docker:  *dockerInfo,
	}, nil
}

func collectSystemMetrics() (*Metrics, error) {
	// CPU
	cpuPercents, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}

	var totalCPU float64
	var cpuInfo []CPUInfo
	for i, percent := range cpuPercents {
		cpuInfo = append(cpuInfo, CPUInfo{
			Name:  fmt.Sprintf("cpu%d", i),
			Usage: percent / 100.0,
		})
		totalCPU += percent
	}

	// Память
	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	swapStat, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}

	// Диск
	diskStats, err := disk.IOCounters()
	if err != nil {
		return nil, err
	}

	var diskInfo []DiskInfo
	for name, stat := range diskStats {
		diskInfo = append(diskInfo, DiskInfo{
			Name:       name,
			ReadBytes:  stat.ReadBytes,
			WriteBytes: stat.WriteBytes,
			Reads:      stat.ReadCount,
			Writes:     stat.WriteCount,
		})
	}

	// Сеть
	netStats, err := gopsutilnet.IOCounters(true)
	if err != nil {
		return nil, err
	}

	var sent, received uint64
	publicIP := getPublicIP()

	for _, stat := range netStats {
		if stat.Name == "eth0" {
			sent = stat.BytesSent
			received = stat.BytesRecv
			break
		}
	}

	return &Metrics{
		CPU: cpuInfo,
		Memory: MemoryInfo{
			RAM: RAMInfo{
				Total: memStat.Total / 1024 / 1024, // MB
				Usage: memStat.Used / 1024 / 1024,  // MB
			},
			Swap: SwapInfo{
				Total: swapStat.Total / 1024 / 1024, // MB
				Usage: swapStat.Used / 1024 / 1024,  // MB
			},
		},
		Disk: diskInfo,
		Network: NetworkInfo{
			PublicIP: publicIP,
			Sent:     sent,
			Received: received,
		},
	}, nil
}

func getPublicIP() string {
	conn, err := net.Dial("udp", "stun.l.google.com:19302")
	if err != nil {
		log.Printf("Failed to connect to STUN server: %v", err)
		return "127.0.0.1"
	}
	defer conn.Close()

	client, err := stun.NewClient(conn)
	if err != nil {
		log.Printf("Failed to create STUN client: %v", err)
		return "127.0.0.1"
	}
	defer client.Close()

	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	var response stun.Event
	if err := client.Do(message, func(res stun.Event) {
		response = res
	}); err != nil {
		log.Printf("Failed to send STUN request: %v", err)
		return "127.0.0.1"
	}

	if response.Error != nil {
		log.Printf("STUN request error: %v", response.Error)
		return "127.0.0.1"
	}

	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(response.Message); err != nil {
		log.Printf("Failed to get XOR mapped address: %v", err)
		return "127.0.0.1"
	}

	return xorAddr.IP.String()
}

func collectDockerMetrics(ctx context.Context, dockerClient *client.Client) (*DockerInfo, error) {
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	containerInfos := []ContainerInfo{}
	for _, container := range containers {
		// Базовая информация о контейнере
		inspect, err := dockerClient.ContainerInspect(ctx, container.ID)
		if err != nil {
			continue
		}

		// Получаем логи
		logs, err := getContainerLogs(ctx, dockerClient, container.ID)
		if err != nil {
			logs = []string{}
		}

		// Получаем статистику контейнера
		stats, err := getContainerStats(ctx, dockerClient, container.ID)
		if err != nil {
			stats = &ContainerStats{}
		}

		networks := []string{}
		var ip, mac *string
		if inspect.NetworkSettings != nil {
			for netName := range inspect.NetworkSettings.Networks {
				networks = append(networks, netName)
			}
			if len(inspect.NetworkSettings.Networks) > 0 {
				for _, net := range inspect.NetworkSettings.Networks {
					if ip == nil {
						if net.IPAddress != "" {
							ip = &net.IPAddress
						}
						if net.MacAddress != "" {
							mac = &net.MacAddress
						}
					}
					break
				}
			}
		}

		containerInfos = append(containerInfos, ContainerInfo{
			ID:           container.ID,
			Created:      time.Unix(container.Created, 0).Format(time.RFC3339Nano),
			Status:       container.Status,
			RestartCount: inspect.RestartCount,
			Image:        strings.TrimPrefix(container.ImageID, "sha256:"),
			Name:         strings.TrimPrefix(container.Names[0], "/"),
			IP:           ip,
			MAC:          mac,
			CPU:          stats.CPU,
			Memory:       stats.Memory,
			Network: ContainerNetworkInfo{
				Sent:     stats.NetworkSent,
				Received: stats.NetworkReceived,
				Networks: networks,
			},
			Logs: logs,
		})
	}

	// Образы
	images, err := dockerClient.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}

	var imageInfos []ImageInfo
	for _, image := range images {
		tags := []string{}
		for _, tag := range image.RepoTags {
			if tag != "<none>:<none>" {
				tags = append(tags, tag)
			}
		}

		imageInfos = append(imageInfos, ImageInfo{
			ID:      strings.TrimPrefix(image.ID, "sha256:"),
			Created: time.Unix(image.Created, 0).Format(time.RFC3339Nano),
			Size:    image.Size,
			Tags:    tags,
		})
	}

	return &DockerInfo{
		Containers: containerInfos,
		Images:     imageInfos,
	}, nil
}

type ContainerStats struct {
	CPU             *float64 `json:"cpu"`
	Memory          *uint64  `json:"memory"`
	NetworkSent     *uint64  `json:"network_sent"`
	NetworkReceived *uint64  `json:"network_received"`
}

func getContainerStats(ctx context.Context, dockerClient *client.Client, containerID string) (*ContainerStats, error) {
	result := &ContainerStats{
		CPU:             nil,
		Memory:          nil,
		NetworkSent:     nil,
		NetworkReceived: nil,
	}

	stats, err := dockerClient.ContainerStatsOneShot(ctx, containerID)
	if err != nil {
		return result, nil
	}

	var v map[string]interface{}
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return result, nil
	}

	// CPU
	cpuUsage := calculateCPUUsage(v)
	if cpuUsage > 0 {
		result.CPU = &cpuUsage
	}

	// Память
	if memStats, ok := v["memory_stats"].(map[string]interface{}); ok {
		if usage, ok := memStats["usage"].(float64); ok && usage > 0 {
			memUsage := uint64(usage) / 1024 / 1024 // MB
			result.Memory = &memUsage
		}
	}

	// Сеть
	if networks, ok := v["networks"].(map[string]interface{}); ok {
		var totalRx, totalTx uint64
		for _, netData := range networks {
			if network, ok := netData.(map[string]interface{}); ok {
				if rxBytes, ok := network["rx_bytes"].(float64); ok {
					totalRx += uint64(rxBytes)
				}
				if txBytes, ok := network["tx_bytes"].(float64); ok {
					totalTx += uint64(txBytes)
				}
			}
		}
		if totalRx > 0 {
			result.NetworkReceived = &totalRx
		}
		if totalTx > 0 {
			result.NetworkSent = &totalTx
		}
	}

	return result, nil
}

func calculateCPUUsage(stats map[string]interface{}) float64 {
	cpuStats, ok := stats["cpu_stats"].(map[string]interface{})
	if !ok {
		return 0
	}

	cpuUsage, ok := cpuStats["cpu_usage"].(map[string]interface{})
	if !ok {
		return 0
	}

	totalUsage, ok := cpuUsage["total_usage"].(float64)
	if !ok {
		return 0
	}

	systemUsage, ok := cpuStats["system_cpu_usage"].(float64)
	if !ok {
		return 0
	}

	var numberCpus float64
	if onlineCpus, ok := cpuStats["online_cpus"].(float64); ok && onlineCpus > 0 {
		numberCpus = onlineCpus
	} else if percpuUsage, ok := cpuUsage["percpu_usage"].([]interface{}); ok && len(percpuUsage) > 0 {
		numberCpus = float64(len(percpuUsage))
	}

	if numberCpus <= 0 {
		return 0
	}

	return (totalUsage / systemUsage) * numberCpus
}

func getContainerLogs(ctx context.Context, dockerClient *client.Client, containerID string) ([]string, error) {
	intervalStr := os.Getenv("INTERVAL")
	if intervalStr == "" {
		intervalStr = "5"
	}

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      intervalStr + "s",
		Timestamps: true,
	}

	logs, err := dockerClient.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return []string{}, err
	}
	defer logs.Close()

	logLines := []string{}
	scanner := bufio.NewScanner(logs)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 8 {
			line = line[8:]
		}
		logLines = append(logLines, line)
	}

	if err := scanner.Err(); err != nil {
		return []string{}, err
	}

	if len(logLines) == 0 {
		return []string{}, nil
	}

	return logLines, nil
}

func sendData(url, token string, data *AgentData) ([]Action, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Парсим ответ с действиями
	var actions []Action
	if err := json.NewDecoder(resp.Body).Decode(&actions); err != nil {
		return nil, fmt.Errorf("failed to decode actions: %v", err)
	}

	return actions, nil
}

// processAction обрабатывает действие от сервера
func processAction(dockerClient *client.Client, action Action) error {
	log.Printf("Processing action %s of type %s", action.ID, action.Type)

	var response *string
	var errMsg *string
	var status string

	switch action.Type {
	case ActionTypeStartContainer:
		response, errMsg, status = handleStartContainer(dockerClient, action.Payload)
	case ActionTypeStopContainer:
		response, errMsg, status = handleStopContainer(dockerClient, action.Payload)
	case ActionTypeRemoveContainer:
		response, errMsg, status = handleRemoveContainer(dockerClient, action.Payload)
	case ActionTypeRemoveImage:
		response, errMsg, status = handleRemoveImage(dockerClient, action.Payload)
	case ActionTypeRestartNginx:
		response, errMsg, status = handleRestartNginx()
	case ActionTypeWriteFile:
		response, errMsg, status = handleWriteFile(action.Payload)
	default:
		err := fmt.Sprintf("Unknown action type: %s", action.Type)
		errMsg = &err
		status = ActionStatusFailed
	}

	// Отправляем результат обратно на сервер
	return sendActionResult(action.ID, status, response, errMsg)
}

// handleStartContainer обрабатывает запуск контейнера
func handleStartContainer(dockerClient *client.Client, payload map[string]interface{}) (*string, *string, string) {
	ctx := context.Background()

	// Извлекаем параметры из payload
	image, ok := payload["image"].(string)
	if !ok {
		err := "Image is required"
		return nil, &err, ActionStatusFailed
	}

	name, ok := payload["name"].(string)
	if !ok {
		err := "Name is required"
		return nil, &err, ActionStatusFailed
	}

	// Создаем конфигурацию контейнера
	containerConfig := &container.Config{
		Image: image,
	}

	// Добавляем порты если указаны
	if ports, ok := payload["ports"].(map[string]interface{}); ok {
		exposedPorts := make(nat.PortSet)
		for port := range ports {
			exposedPorts[nat.Port(port)] = struct{}{}
		}
		containerConfig.ExposedPorts = exposedPorts
	}

	// Добавляем переменные окружения если указаны
	if env, ok := payload["environment"].(map[string]interface{}); ok {
		for key, value := range env {
			if strValue, ok := value.(string); ok {
				containerConfig.Env = append(containerConfig.Env, fmt.Sprintf("%s=%s", key, strValue))
			}
		}
	}

	// Создаем конфигурацию хоста
	hostConfig := &container.HostConfig{}

	// Добавляем порты если указаны
	if ports, ok := payload["ports"].(map[string]interface{}); ok {
		portBindings := make(nat.PortMap)
		for containerPort, hostPort := range ports {
			if hostPortStr, ok := hostPort.(string); ok {
				portBindings[nat.Port(containerPort)] = []nat.PortBinding{
					{HostPort: hostPortStr},
				}
			}
		}
		hostConfig.PortBindings = portBindings
	}

	// Добавляем volumes если указаны
	if volumes, ok := payload["volumes"].(map[string]interface{}); ok {
		binds := make([]string, 0)
		for hostPath, containerPath := range volumes {
			if containerPathStr, ok := containerPath.(string); ok {
				binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPathStr))
			}
		}
		hostConfig.Binds = binds
	}

	// Создаем контейнер
	resp, err := dockerClient.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, name)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create container: %v", err)
		return nil, &errMsg, ActionStatusFailed
	}

	// Запускаем контейнер
	err = dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to start container: %v", err)
		return nil, &errMsg, ActionStatusFailed
	}

	// Если указан домен, обновляем nginx конфигурацию
	if domain, ok := payload["domain"].(string); ok && domain != "" {
		if err := updateNginxConfig(domain, name); err != nil {
			log.Printf("Warning: failed to update nginx config: %v", err)
		}
	}

	successMsg := fmt.Sprintf("Container %s started successfully with ID: %s", name, resp.ID)
	return &successMsg, nil, ActionStatusCompleted
}

// handleStopContainer обрабатывает остановку контейнера
func handleStopContainer(dockerClient *client.Client, payload map[string]interface{}) (*string, *string, string) {
	ctx := context.Background()

	containerID, ok := payload["container_id"].(string)
	if !ok {
		err := "Container ID is required"
		return nil, &err, ActionStatusFailed
	}

	timeout := 10
	if timeoutVal, ok := payload["timeout"].(float64); ok {
		timeout = int(timeoutVal)
	}

	// Останавливаем контейнер
	err := dockerClient.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to stop container: %v", err)
		return nil, &errMsg, ActionStatusFailed
	}

	successMsg := fmt.Sprintf("Container %s stopped successfully", containerID)
	return &successMsg, nil, ActionStatusCompleted
}

// handleRemoveContainer обрабатывает удаление контейнера
func handleRemoveContainer(dockerClient *client.Client, payload map[string]interface{}) (*string, *string, string) {
	ctx := context.Background()

	containerID, ok := payload["container_id"].(string)
	if !ok {
		err := "Container ID is required"
		return nil, &err, ActionStatusFailed
	}

	force := false
	if forceVal, ok := payload["force"].(bool); ok {
		force = forceVal
	}

	// Удаляем контейнер
	err := dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: force,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to remove container: %v", err)
		return nil, &errMsg, ActionStatusFailed
	}

	successMsg := fmt.Sprintf("Container %s removed successfully", containerID)
	return &successMsg, nil, ActionStatusCompleted
}

// handleRemoveImage обрабатывает удаление образа
func handleRemoveImage(dockerClient *client.Client, payload map[string]interface{}) (*string, *string, string) {
	ctx := context.Background()

	imageID, ok := payload["image_id"].(string)
	if !ok {
		err := "Image ID is required"
		return nil, &err, ActionStatusFailed
	}

	force := false
	if forceVal, ok := payload["force"].(bool); ok {
		force = forceVal
	}

	// Удаляем образ
	_, err := dockerClient.ImageRemove(ctx, imageID, image.RemoveOptions{
		Force: force,
	})
	if err != nil {
		errMsg := fmt.Sprintf("Failed to remove image: %v", err)
		return nil, &errMsg, ActionStatusFailed
	}

	successMsg := fmt.Sprintf("Image %s removed successfully", imageID)
	return &successMsg, nil, ActionStatusCompleted
}

// handleRestartNginx обрабатывает перезапуск nginx
func handleRestartNginx() (*string, *string, string) {
	// Выполняем команду для перезапуска nginx
	cmd := exec.Command("systemctl", "restart", "nginx")
	if err := cmd.Run(); err != nil {
		errMsg := fmt.Sprintf("Failed to restart nginx: %v", err)
		return nil, &errMsg, ActionStatusFailed
	}

	successMsg := "Nginx restarted successfully"
	return &successMsg, nil, ActionStatusCompleted
}

// handleWriteFile обрабатывает запись в файл
func handleWriteFile(payload map[string]interface{}) (*string, *string, string) {
	path, ok := payload["path"].(string)
	if !ok {
		err := "Path is required"
		return nil, &err, ActionStatusFailed
	}

	content, ok := payload["content"].(string)
	if !ok {
		err := "Content is required"
		return nil, &err, ActionStatusFailed
	}

	mode := 0644
	if modeVal, ok := payload["mode"].(float64); ok {
		mode = int(modeVal)
	}

	// Создаем директорию если не существует
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		errMsg := fmt.Sprintf("Failed to create directory: %v", err)
		return nil, &errMsg, ActionStatusFailed
	}

	// Записываем файл
	if err := os.WriteFile(path, []byte(content), os.FileMode(mode)); err != nil {
		errMsg := fmt.Sprintf("Failed to write file: %v", err)
		return nil, &errMsg, ActionStatusFailed
	}

	successMsg := fmt.Sprintf("File %s written successfully", path)
	return &successMsg, nil, ActionStatusCompleted
}

// updateNginxConfig обновляет конфигурацию nginx для нового домена
func updateNginxConfig(domain, containerName string) error {
	// Создаем конфигурацию nginx для домена
	config := fmt.Sprintf(`
server {
    listen 80;
    server_name %s;
    
    location / {
        proxy_pass http://%s;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
`, domain, containerName)

	// Записываем конфигурацию в файл
	configPath := fmt.Sprintf("/etc/nginx/sites-available/%s", domain)
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return err
	}

	// Создаем символическую ссылку
	symlinkPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", domain)
	if err := os.Symlink(configPath, symlinkPath); err != nil && !os.IsExist(err) {
		return err
	}

	// Проверяем конфигурацию nginx
	cmd := exec.Command("nginx", "-t")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Перезапускаем nginx
	cmd = exec.Command("systemctl", "reload", "nginx")
	return cmd.Run()
}

// sendActionResult отправляет результат выполнения действия на сервер
func sendActionResult(actionID, status string, response, error *string) error {
	url := os.Getenv("URL")
	token := os.Getenv("TOKEN")

	// Формируем URL для обновления статуса действия
	updateURL := strings.TrimSuffix(url, "/agent/ping") + "/actions/" + actionID + "/status"

	// Создаем ответ
	result := ActionResponse{
		ID:       actionID,
		Status:   status,
		Response: response,
		Error:    error,
	}

	// Сериализуем ответ
	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// Создаем запрос
	req, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Отправляем запрос
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

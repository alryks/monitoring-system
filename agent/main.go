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

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/pion/stun"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	gopsutilnet "github.com/shirou/gopsutil/v3/net"
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
	Volumes    []VolumeInfo    `json:"volumes"`
	Networks   []NetworkDocker `json:"networks"`
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
	Volumes      []string             `json:"volumes"`
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

type VolumeInfo struct {
	Name       string `json:"name"`
	Created    string `json:"created"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint"`
}

type NetworkDocker struct {
	ID      string  `json:"id"`
	Created string  `json:"created"`
	Name    string  `json:"name"`
	Driver  string  `json:"driver"`
	Scope   string  `json:"scope"`
	Subnet  *string `json:"subnet"`
	Gateway *string `json:"gateway"`
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
			err = sendData(url, token, data)
			if err != nil {
				log.Printf("Error sending data: %v", err)
			} else {
				log.Println("Data sent successfully")
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

		volumes := []string{}
		for _, mount := range inspect.Mounts {
			if mount.Type == "volume" {
				volumes = append(volumes, mount.Name)
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
			Volumes: volumes,
			Logs:    logs,
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

	// Тома
	volumeList, err := dockerClient.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return nil, err
	}

	var volumeInfos []VolumeInfo
	for _, volume := range volumeList.Volumes {
		volumeInfos = append(volumeInfos, VolumeInfo{
			Name:       volume.Name,
			Created:    volume.CreatedAt,
			Driver:     volume.Driver,
			Mountpoint: volume.Mountpoint,
		})
	}

	// Сети
	networks, err := dockerClient.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	var networkInfos []NetworkDocker
	for _, network := range networks {
		var subnet, gateway *string
		if len(network.IPAM.Config) > 0 {
			if network.IPAM.Config[0].Subnet != "" {
				subnet = &network.IPAM.Config[0].Subnet
			}
			if network.IPAM.Config[0].Gateway != "" {
				gateway = &network.IPAM.Config[0].Gateway
			}
		}

		networkInfos = append(networkInfos, NetworkDocker{
			ID:      network.ID,
			Created: network.Created.Format(time.RFC3339Nano),
			Name:    network.Name,
			Driver:  network.Driver,
			Scope:   network.Scope,
			Subnet:  subnet,
			Gateway: gateway,
		})
	}

	return &DockerInfo{
		Containers: containerInfos,
		Images:     imageInfos,
		Volumes:    volumeInfos,
		Networks:   networkInfos,
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

func sendData(url, token string, data *AgentData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

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

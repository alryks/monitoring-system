package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"monitoring-system/core/server/internal/models"
)

// Service представляет сервис для работы с уведомлениями
type Service struct {
	settings *models.NotificationSettings
	client   *http.Client
}

// New создает новый экземпляр сервиса уведомлений
func New() *Service {
	return &Service{
		settings: &models.NotificationSettings{
			TelegramBotToken: "",
			TelegramChatID:   "",
			Notifications: models.NotificationConfigurations{
				AgentOffline: models.NotificationConfig{
					Enabled: false,
					Message: "🚨 Агент {AGENT_NAME} не отвечает!",
				},
				ContainerStopped: models.NotificationConfig{
					Enabled: false,
					Message: "⚠️ Контейнер {CONTAINER_NAME} остановился на агенте {AGENT_NAME}",
				},
				CPUThreshold: models.CPUThresholdConfig{
					Enabled:   false,
					Threshold: 80,
					Message:   "🔥 Высокое использование CPU: {AGENT_NAME} - {CPU_USAGE}%",
				},
				RAMThreshold: models.RAMThresholdConfig{
					Enabled:   false,
					Threshold: 80,
					Message:   "💾 Высокое использование RAM: {AGENT_NAME} - {RAM_USAGE}%",
				},
			},
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetSettings возвращает текущие настройки уведомлений
func (s *Service) GetSettings() *models.NotificationSettings {
	return s.settings
}

// UpdateSettings обновляет настройки уведомлений
func (s *Service) UpdateSettings(settings *models.NotificationSettings) {
	s.settings = settings
}

// SendTestNotification отправляет тестовое уведомление
func (s *Service) SendTestNotification() error {
	if s.settings.TelegramBotToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	message := "🧪 Тестовое уведомление от системы мониторинга\n\nВремя: " + time.Now().Format("2006-01-02 15:04:05")

	return s.sendTelegramMessage(message)
}

// CheckAgentOffline проверяет, нужно ли отправить уведомление о недоступности агента
func (s *Service) CheckAgentOffline(agentName string) error {
	if !s.settings.Notifications.AgentOffline.Enabled {
		return nil
	}

	message := s.replaceVariables(s.settings.Notifications.AgentOffline.Message, map[string]string{
		"AGENT_NAME": agentName,
	})

	return s.sendTelegramMessage(message)
}

// CheckContainerStopped проверяет, нужно ли отправить уведомление об остановке контейнера
func (s *Service) CheckContainerStopped(containerName, agentName string) error {
	if !s.settings.Notifications.ContainerStopped.Enabled {
		return nil
	}

	message := s.replaceVariables(s.settings.Notifications.ContainerStopped.Message, map[string]string{
		"CONTAINER_NAME": containerName,
		"AGENT_NAME":     agentName,
	})

	return s.sendTelegramMessage(message)
}

// CheckCPUThreshold проверяет, нужно ли отправить уведомление о превышении CPU
func (s *Service) CheckCPUThreshold(agentName string, cpuUsage float64) error {
	if !s.settings.Notifications.CPUThreshold.Enabled {
		return nil
	}

	if cpuUsage > float64(s.settings.Notifications.CPUThreshold.Threshold) {
		message := s.replaceVariables(s.settings.Notifications.CPUThreshold.Message, map[string]string{
			"AGENT_NAME": agentName,
			"CPU_USAGE":  fmt.Sprintf("%.1f", cpuUsage),
		})

		return s.sendTelegramMessage(message)
	}

	return nil
}

// CheckRAMThreshold проверяет, нужно ли отправить уведомление о превышении RAM
func (s *Service) CheckRAMThreshold(agentName string, ramUsage float64) error {
	if !s.settings.Notifications.RAMThreshold.Enabled {
		return nil
	}

	if ramUsage > float64(s.settings.Notifications.RAMThreshold.Threshold) {
		message := s.replaceVariables(s.settings.Notifications.RAMThreshold.Message, map[string]string{
			"AGENT_NAME": agentName,
			"RAM_USAGE":  fmt.Sprintf("%.1f", ramUsage),
		})

		return s.sendTelegramMessage(message)
	}

	return nil
}

// sendTelegramMessage отправляет сообщение в Telegram
func (s *Service) sendTelegramMessage(text string) error {
	if s.settings.TelegramBotToken == "" {
		return fmt.Errorf("telegram bot token not configured")
	}

	if s.settings.TelegramChatID == "" {
		return fmt.Errorf("telegram chat ID not configured")
	}

	message := models.TelegramMessage{
		ChatID:    s.settings.TelegramChatID,
		Text:      text,
		ParseMode: "HTML",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.settings.TelegramBotToken)

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending telegram message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Читаем тело ответа для получения деталей ошибки
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("Telegram notification sent successfully")
	return nil
}

// replaceVariables заменяет переменные в сообщении
func (s *Service) replaceVariables(message string, variables map[string]string) string {
	result := message
	for key, value := range variables {
		result = strings.ReplaceAll(result, "{"+key+"}", value)
	}
	return result
}

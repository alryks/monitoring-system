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

	"gopkg.in/gomail.v2"
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
			EmailSettings: models.EmailSettings{
				Enabled:     false,
				SMTPHost:    "",
				SMTPPort:    587,
				Username:    "",
				Password:    "",
				FromEmail:   "",
				FromName:    "Система мониторинга",
				ToEmails:    "",
				UseTLS:      true,
				UseStartTLS: true,
			},
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
	message := "🧪 Тестовое уведомление от системы мониторинга\n\nВремя: " + time.Now().Format("2006-01-02 15:04:05")

	// Отправляем в Telegram
	if s.settings.TelegramBotToken != "" && s.settings.TelegramChatID != "" {
		if err := s.sendTelegramMessage(message); err != nil {
			log.Printf("Error sending Telegram test notification: %v", err)
		}
	}

	// Отправляем email
	if s.settings.EmailSettings.Enabled {
		if err := s.sendEmailMessage("Тестовое уведомление", message); err != nil {
			log.Printf("Error sending email test notification: %v", err)
			return err
		}
	}

	return nil
}

// CheckAgentOffline проверяет, нужно ли отправить уведомление о недоступности агента
func (s *Service) CheckAgentOffline(agentName string) error {
	if !s.settings.Notifications.AgentOffline.Enabled {
		return nil
	}

	message := s.replaceVariables(s.settings.Notifications.AgentOffline.Message, map[string]string{
		"AGENT_NAME": agentName,
	})

	// Отправляем в Telegram
	if s.settings.TelegramBotToken != "" && s.settings.TelegramChatID != "" {
		if err := s.sendTelegramMessage(message); err != nil {
			log.Printf("Error sending Telegram agent offline notification: %v", err)
		}
	}

	// Отправляем email
	if s.settings.EmailSettings.Enabled {
		if err := s.sendEmailMessage("Агент не отвечает", message); err != nil {
			log.Printf("Error sending email agent offline notification: %v", err)
			return err
		}
	}

	return nil
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

	// Отправляем в Telegram
	if s.settings.TelegramBotToken != "" && s.settings.TelegramChatID != "" {
		if err := s.sendTelegramMessage(message); err != nil {
			log.Printf("Error sending Telegram container stopped notification: %v", err)
		}
	}

	// Отправляем email
	if s.settings.EmailSettings.Enabled {
		if err := s.sendEmailMessage("Контейнер остановился", message); err != nil {
			log.Printf("Error sending email container stopped notification: %v", err)
			return err
		}
	}

	return nil
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

		// Отправляем в Telegram
		if s.settings.TelegramBotToken != "" && s.settings.TelegramChatID != "" {
			if err := s.sendTelegramMessage(message); err != nil {
				log.Printf("Error sending Telegram CPU threshold notification: %v", err)
			}
		}

		// Отправляем email
		if s.settings.EmailSettings.Enabled {
			if err := s.sendEmailMessage("Высокое использование CPU", message); err != nil {
				log.Printf("Error sending email CPU threshold notification: %v", err)
				return err
			}
		}
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

		// Отправляем в Telegram
		if s.settings.TelegramBotToken != "" && s.settings.TelegramChatID != "" {
			if err := s.sendTelegramMessage(message); err != nil {
				log.Printf("Error sending Telegram RAM threshold notification: %v", err)
			}
		}

		// Отправляем email
		if s.settings.EmailSettings.Enabled {
			if err := s.sendEmailMessage("Высокое использование RAM", message); err != nil {
				log.Printf("Error sending email RAM threshold notification: %v", err)
				return err
			}
		}
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

// sendEmailMessage отправляет email сообщение
func (s *Service) sendEmailMessage(subject, body string) error {
	if !s.settings.EmailSettings.Enabled {
		return fmt.Errorf("email notifications not enabled")
	}

	if s.settings.EmailSettings.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	if s.settings.EmailSettings.ToEmails == "" {
		return fmt.Errorf("recipient emails not configured")
	}

	// Создаем сообщение
	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.settings.EmailSettings.FromEmail, s.settings.EmailSettings.FromName)
	
	// Разбираем список получателей
	toEmails := strings.Split(s.settings.EmailSettings.ToEmails, ",")
	for i, email := range toEmails {
		toEmails[i] = strings.TrimSpace(email)
	}
	m.SetHeader("To", toEmails...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Настраиваем SMTP
	d := gomail.NewDialer(
		s.settings.EmailSettings.SMTPHost,
		s.settings.EmailSettings.SMTPPort,
		s.settings.EmailSettings.Username,
		s.settings.EmailSettings.Password,
	)

	if s.settings.EmailSettings.UseTLS {
		d.SSL = true
	} else if s.settings.EmailSettings.UseStartTLS {
		d.TLSConfig = nil // gomail автоматически использует STARTTLS
	}

	// Отправляем сообщение
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	log.Printf("Email notification sent successfully to %s", s.settings.EmailSettings.ToEmails)
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

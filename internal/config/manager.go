package config

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

type RequestIDConfig struct {
	RequireRequestID        bool   `yaml:"require_request_id"`
	OverrideServiceSettings bool   `yaml:"override_service_settings"`
	ErrorMessage            string `yaml:"error_message"`
}

type APIKeyConfig struct {
	APIKey                  string `yaml:"api_key"`
	RequireAPIKey           bool   `yaml:"require_api_key"`
	OverrideServiceSettings bool   `yaml:"override_service_settings"`
	ErrorMessage            string `yaml:"error_message"`
}

type Config struct {
	Message         string           `yaml:"message"`
	RequestID       RequestIDConfig  `yaml:"request_id"`
	APIKey          APIKeyConfig     `yaml:"api_key"`
	BackendServices []BackendService `yaml:"backend_services"`
}

type BackendService struct {
	Name             string `yaml:"name"`
	URL              string `yaml:"url"`
	Enabled          bool   `yaml:"enabled"`
	RequireRequestID *bool  `yaml:"require_request_id,omitempty"`
	RequireAPIKey    *bool  `yaml:"require_api_key,omitempty"`
}

type Manager struct {
	config      *Config
	mu          sync.RWMutex
	watcher     *fsnotify.Watcher
	configPath  string
	lastModTime time.Time
}

func NewManager(configPath string) (*Manager, error) {
	manager := &Manager{
		configPath: configPath,
		config:     &Config{},
	}

	// load initial config
	if err := manager.LoadConfig(); err != nil {
		log.Printf("warning: failed to load initial config: %v", err)
	}

	// setup file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	manager.watcher = watcher

	return manager, nil
}

func (m *Manager) LoadConfig() error {
	fileInfo, err := os.Stat(m.configPath)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var newConfig Config
	if err := yaml.Unmarshal(data, &newConfig); err != nil {
		return err
	}

	if newConfig.RequestID.ErrorMessage == "" {
		newConfig.RequestID.ErrorMessage = "Missing required header: x-request-id"
	}

	if newConfig.APIKey.ErrorMessage == "" {
		newConfig.APIKey.ErrorMessage = "Missing required header: x-api-key"
	}

	m.mu.Lock()
	m.config = &newConfig
	m.lastModTime = fileInfo.ModTime()
	m.mu.Unlock()

	log.Printf("config loaded: message='%s'", newConfig.Message)
	return nil
}

func (m *Manager) GetMessage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config.Message == "" {
		return "pong"
	}
	return m.config.Message
}

func (m *Manager) GetBackendServices() []BackendService {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config.BackendServices
}

func (m *Manager) GetBackendService(name string) (*BackendService, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, service := range m.config.BackendServices {
		if service.Name == name && service.Enabled {
			return &service, true
		}
	}
	return nil, false
}

func (m *Manager) StartWatching() error {
	if err := m.watcher.Add(m.configPath); err != nil {
		return err
	}

	// start fsnotify watcher
	go func() {
		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					log.Printf("config file modified via fsnotify: %s", event.Name)
					if err := m.LoadConfig(); err != nil {
						log.Printf("error reloading config: %v", err)
					}
				}
			case err, ok := <-m.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("config watcher error: %v", err)
			}
		}
	}()

	// start polling as fallback (useful for Docker volumes on macOS)
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			fileInfo, err := os.Stat(m.configPath)
			if err != nil {
				log.Printf("error checking config file: %v", err)
				continue
			}

			m.mu.RLock()
			lastMod := m.lastModTime
			m.mu.RUnlock()

			if fileInfo.ModTime().After(lastMod) {
				log.Printf("config file modified via polling: %s", m.configPath)
				if err := m.LoadConfig(); err != nil {
					log.Printf("error reloading config: %v", err)
				}
			}
		}
	}()

	log.Printf("started watching config file: %s (fsnotify + polling)", m.configPath)
	return nil
}

func (m *Manager) Stop() error {
	if m.watcher != nil {
		return m.watcher.Close()
	}
	return nil
}

func (m *Manager) GetRequestIDConfig() RequestIDConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.RequestID
}

func (m *Manager) ShouldRequireRequestID(serviceName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	globalConfig := m.config.RequestID

	// si override_service_settings está en true, usar solo el global
	if globalConfig.OverrideServiceSettings {
		return globalConfig.RequireRequestID
	}

	// buscar configuración específica del service
	for _, service := range m.config.BackendServices {
		if service.Name == serviceName {
			// si el service tiene configuración específica, usarla
			if service.RequireRequestID != nil {
				return *service.RequireRequestID
			}
			// si no tiene configuración específica, usar global como default
			return globalConfig.RequireRequestID
		}
	}

	// si no se encuentra el service, usar global como fallback
	return globalConfig.RequireRequestID
}

func (m *Manager) GetRequestIDErrorMessage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config.RequestID.ErrorMessage == "" {
		return "Missing required header: x-request-id"
	}
	return m.config.RequestID.ErrorMessage
}

func (m *Manager) GetAPIKeyConfig() APIKeyConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.APIKey
}

func (m *Manager) ShouldRequireAPIKey(serviceName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	globalConfig := m.config.APIKey

	// si override_service_settings está en true, usar solo el global
	if globalConfig.OverrideServiceSettings {
		return globalConfig.RequireAPIKey
	}

	// buscar configuración específica del service
	for _, service := range m.config.BackendServices {
		if service.Name == serviceName {
			// si el service tiene configuración específica, usarla
			if service.RequireAPIKey != nil {
				return *service.RequireAPIKey
			}
			// si no tiene configuración específica, usar global como default
			return globalConfig.RequireAPIKey
		}
	}

	// si no se encuentra el service, usar global como fallback
	return globalConfig.RequireAPIKey
}

func (m *Manager) GetAPIKeyErrorMessage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config.APIKey.ErrorMessage == "" {
		return "Missing required header: x-api-key"
	}
	return m.config.APIKey.ErrorMessage
}

func (m *Manager) IsValidAPIKey(providedKey string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config.APIKey.APIKey != "" && m.config.APIKey.APIKey == providedKey
}

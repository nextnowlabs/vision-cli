package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Record struct {
	ID           string   `json:"id"`
	Timestamp    string   `json:"timestamp"`
	Prompt       string   `json:"prompt"`
	InputImages  []string `json:"input_images"`
	OutputImages []string `json:"output_images"`
	AspectRatio  string   `json:"aspect_ratio,omitempty"`
	Resolution   string   `json:"resolution,omitempty"`
	Mode         string   `json:"mode"`
	Status       string   `json:"status"`
	Model        string   `json:"model"`
	Backend      string   `json:"backend"`
	VideoTaskID  string   `json:"video_task_id,omitempty"`
	Duration     int      `json:"duration,omitempty"`
}

type Stats struct {
	TotalCalls  int
	Success     int
	Failed      int
	Direct      int
	Batch       int
	Video       int
	TotalImages int
	Monthly     map[string]int
	Daily       map[string]int
}

type Store struct {
	mu      sync.RWMutex
	dataDir string
}

func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &Store{dataDir: dataDir}, nil
}

func configDir() (string, error) {
	if runtime.GOOS == "windows" {
		dir := os.Getenv("APPDATA")
		if dir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, "AppData", "Roaming", "vision-cli"), nil
		}
		return filepath.Join(dir, "vision-cli"), nil
	}

	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "vision-cli"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "vision-cli"), nil
	}
	return filepath.Join(home, ".config", "vision-cli"), nil
}

func DefaultStore() (*Store, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	return NewStore(dir)
}

func randomID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func nowISO() string {
	return time.Now().Format(time.RFC3339)
}

func (s *Store) configPath() string {
	return filepath.Join(s.dataDir, "config.json")
}

func (s *Store) historyPath() string {
	return filepath.Join(s.dataDir, "history.json")
}

func (s *Store) LoadConfig() (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadConfig()
}

func (s *Store) loadConfig() (map[string]any, error) {
	cfg := make(map[string]any)
	data, err := os.ReadFile(s.configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Store) SetConfig(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadConfig()
	if err != nil {
		return err
	}
	cfg[key] = value
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath(), data, 0644)
}

func (s *Store) GetAPIKey(backend string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, err := s.loadConfig()
	if err != nil {
		cfg = make(map[string]any)
	}

	var configKey, envVar string
	switch backend {
	case "volcengine_ark":
		configKey = "ark_api_key"
		envVar = "ARK_API_KEY"
	default:
		return "", false
	}

	if v, ok := cfg[configKey]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s, true
		}
	}

	if v := os.Getenv(envVar); v != "" {
		return v, true
	}

	return "", false
}

func (s *Store) AddRecord(r Record) (string, error) {
	id, err := randomID()
	if err != nil {
		return "", err
	}
	r.ID = id
	r.Timestamp = nowISO()

	s.mu.Lock()
	defer s.mu.Unlock()

	records, err := s.readHistory()
	if err != nil {
		return "", err
	}

	records = append(records, r)
	if err := s.writeHistory(records); err != nil {
		return "", err
	}

	return r.ID, nil
}

func (s *Store) GetRecords(limit int, search string) ([]Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, err := s.readHistory()
	if err != nil {
		return nil, err
	}

	var filtered []Record
	for _, r := range records {
		if search != "" && !strings.Contains(strings.ToLower(r.Prompt), strings.ToLower(search)) {
			continue
		}
		filtered = append(filtered, r)
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[len(filtered)-limit:]
	}

	return filtered, nil
}

func (s *Store) GetRecord(id string) (*Record, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, err := s.readHistory()
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.ID == id {
			return &r, nil
		}
	}

	return nil, errors.New("record not found")
}

func (s *Store) GetStats() (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, err := s.readHistory()
	if err != nil {
		return nil, err
	}

	stats := &Stats{
		Monthly: make(map[string]int),
		Daily:   make(map[string]int),
	}

	stats.TotalCalls = len(records)
	for _, r := range records {
		switch r.Status {
		case "success":
			stats.Success++
		case "failed":
			stats.Failed++
		}

		switch r.Mode {
		case "direct":
			stats.Direct++
		case "batch":
			stats.Batch++
		case "video":
			stats.Video++
		}

		stats.TotalImages += len(r.OutputImages)

		if r.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, r.Timestamp); err == nil {
				stats.Monthly[t.Format("2006-01")]++
				stats.Daily[t.Format("2006-01-02")]++
			}
		}
	}

	return stats, nil
}

func (s *Store) readHistory() ([]Record, error) {
	var records []Record
	data, err := os.ReadFile(s.historyPath())
	if err != nil {
		if os.IsNotExist(err) {
			return records, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *Store) writeHistory(records []Record) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.historyPath(), data, 0644)
}



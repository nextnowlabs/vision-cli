package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	return s
}

func TestLoadConfig_Empty(t *testing.T) {
	s := tempStore(t)
	cfg, err := s.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if len(cfg) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(cfg))
	}
}

func TestSetConfig_LoadConfig(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if err := s.SetConfig("key1", "value1"); err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}
	if err := s.SetConfig("key2", float64(42)); err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}

	cfg, err := s.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if v, ok := cfg["key1"]; !ok || v != "value1" {
		t.Fatalf("key1 = %v, want 'value1'", v)
	}
	if v, ok := cfg["key2"]; !ok || v != float64(42) {
		t.Fatalf("key2 = %v, want 42", v)
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("reading config.json: %v", err)
	}
	var onDisk map[string]any
	if err := json.Unmarshal(data, &onDisk); err != nil {
		t.Fatalf("unmarshaling config.json: %v", err)
	}
	if onDisk["key1"] != "value1" || onDisk["key2"] != float64(42) {
		t.Fatalf("on-disk mismatch: %v", onDisk)
	}
}

func TestGetAPIKey_ConfigOnly(t *testing.T) {
	s := tempStore(t)
	s.SetConfig("dashscope_api_key", "sk-cfg-scope")

	key, ok := s.GetAPIKey("dashscope")
	if !ok || key != "sk-cfg-scope" {
		t.Fatalf("GetAPIKey(dashscope) = (%q, %v), want (sk-cfg-scope, true)", key, ok)
	}
}

func TestGetAPIKey_EnvVar(t *testing.T) {
	s := tempStore(t)

	os.Setenv("DASHSCOPE_API_KEY", "sk-env-scope")
	defer os.Unsetenv("DASHSCOPE_API_KEY")

	key, ok := s.GetAPIKey("dashscope")
	if !ok || key != "sk-env-scope" {
		t.Fatalf("GetAPIKey(dashscope) = (%q, %v), want (sk-env-scope, true)", key, ok)
	}

	os.Setenv("ARK_API_KEY", "sk-env-ark")
	defer os.Unsetenv("ARK_API_KEY")

	key, ok = s.GetAPIKey("volcengine_ark")
	if !ok || key != "sk-env-ark" {
		t.Fatalf("GetAPIKey(volcengine_ark) = (%q, %v), want (sk-env-ark, true)", key, ok)
	}
}

func TestGetAPIKey_ConfigOverridesEnv(t *testing.T) {
	s := tempStore(t)
	s.SetConfig("dashscope_api_key", "sk-cfg-override")
	os.Setenv("DASHSCOPE_API_KEY", "sk-env-orig")
	defer os.Unsetenv("DASHSCOPE_API_KEY")

	key, ok := s.GetAPIKey("dashscope")
	if !ok || key != "sk-cfg-override" {
		t.Fatalf("GetAPIKey(dashscope) = (%q, %v), want (sk-cfg-override, true)", key, ok)
	}
}

func TestGetAPIKey_Missing(t *testing.T) {
	s := tempStore(t)

	_, ok := s.GetAPIKey("dashscope")
	if ok {
		t.Fatal("expected false for missing key")
	}

	_, ok = s.GetAPIKey("unknown")
	if ok {
		t.Fatal("expected false for unknown backend")
	}
}

func TestAddRecord_GetRecords(t *testing.T) {
	s := tempStore(t)

	r := Record{
		Prompt:  "a beautiful sunset",
		Mode:    "direct",
		Status:  "success",
		Model:   "flux",
		Backend: "dashscope",
	}
	id, err := s.AddRecord(r)
	if err != nil {
		t.Fatalf("AddRecord() error = %v", err)
	}
	if len(id) != 8 {
		t.Fatalf("expected 8-char ID, got %q", id)
	}

	records, err := s.GetRecords(0, "")
	if err != nil {
		t.Fatalf("GetRecords() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Prompt != "a beautiful sunset" {
		t.Fatalf("prompt = %q", records[0].Prompt)
	}
	if records[0].ID != id {
		t.Fatalf("ID = %q, want %q", records[0].ID, id)
	}
}

func TestGetRecords_Search(t *testing.T) {
	s := tempStore(t)

	s.AddRecord(Record{Prompt: "hello world", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "foo bar", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})

	records, err := s.GetRecords(0, "hello")
	if err != nil {
		t.Fatalf("GetRecords() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for search 'hello', got %d", len(records))
	}
	if records[0].Prompt != "hello world" {
		t.Fatalf("prompt = %q", records[0].Prompt)
	}

	records, err = s.GetRecords(0, "zzz")
	if err != nil {
		t.Fatalf("GetRecords() error = %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records for search 'zzz', got %d", len(records))
	}
}

func TestGetRecords_Limit(t *testing.T) {
	s := tempStore(t)

	s.AddRecord(Record{Prompt: "a", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "b", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "c", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})

	records, err := s.GetRecords(2, "")
	if err != nil {
		t.Fatalf("GetRecords() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records with limit, got %d", len(records))
	}
}

func TestGetRecords_SearchCaseInsensitive(t *testing.T) {
	s := tempStore(t)

	s.AddRecord(Record{Prompt: "a cat on a mat", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "a dog running", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})

	records, err := s.GetRecords(0, "CAT")
	if err != nil {
		t.Fatalf("GetRecords() error = %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for case-insensitive search 'CAT', got %d", len(records))
	}
	if records[0].Prompt != "a cat on a mat" {
		t.Fatalf("prompt = %q, want 'a cat on a mat'", records[0].Prompt)
	}
}

func TestGetRecords_LimitReturnsNewest(t *testing.T) {
	s := tempStore(t)

	s.AddRecord(Record{Prompt: "first", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "second", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "third", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "fourth", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "fifth", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})

	records, err := s.GetRecords(2, "")
	if err != nil {
		t.Fatalf("GetRecords() error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records with limit, got %d", len(records))
	}
	if records[0].Prompt != "fourth" {
		t.Fatalf("first record prompt = %q, want 'fourth'", records[0].Prompt)
	}
	if records[1].Prompt != "fifth" {
		t.Fatalf("second record prompt = %q, want 'fifth'", records[1].Prompt)
	}
}

func TestGetRecord_ByID(t *testing.T) {
	s := tempStore(t)

	id, _ := s.AddRecord(Record{Prompt: "test", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope"})

	rec, err := s.GetRecord(id)
	if err != nil {
		t.Fatalf("GetRecord() error = %v", err)
	}
	if rec.ID != id || rec.Prompt != "test" {
		t.Fatalf("record mismatch: %+v", rec)
	}
}

func TestGetRecord_NotFound(t *testing.T) {
	s := tempStore(t)

	_, err := s.GetRecord("deadbeef")
	if err == nil {
		t.Fatal("expected error for non-existent record")
	}
}

func TestGetStats(t *testing.T) {
	s := tempStore(t)

	s.AddRecord(Record{Prompt: "p1", Mode: "direct", Status: "success", Model: "flux", Backend: "dashscope", OutputImages: []string{"a.png", "b.png"}})
	s.AddRecord(Record{Prompt: "p2", Mode: "batch", Status: "success", Model: "flux", Backend: "dashscope", OutputImages: []string{"c.png"}})
	s.AddRecord(Record{Prompt: "p3", Mode: "video", Status: "success", Model: "flux", Backend: "dashscope"})
	s.AddRecord(Record{Prompt: "p4", Mode: "direct", Status: "failed", Model: "flux", Backend: "dashscope"})

	stats, err := s.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalCalls != 4 {
		t.Fatalf("TotalCalls = %d, want 4", stats.TotalCalls)
	}
	if stats.Success != 3 {
		t.Fatalf("Success = %d, want 3", stats.Success)
	}
	if stats.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", stats.Failed)
	}
	if stats.Direct != 2 {
		t.Fatalf("Direct = %d, want 2", stats.Direct)
	}
	if stats.Batch != 1 {
		t.Fatalf("Batch = %d, want 1", stats.Batch)
	}
	if stats.Video != 1 {
		t.Fatalf("Video = %d, want 1", stats.Video)
	}
	if stats.TotalImages != 3 {
		t.Fatalf("TotalImages = %d, want 3", stats.TotalImages)
	}
}

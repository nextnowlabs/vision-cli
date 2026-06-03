package client

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackendOf(t *testing.T) {
	tests := []struct {
		alias string
		want  string
	}{
		{"seedream", "volcengine_ark"},
		{"seedream-lite", "volcengine_ark"},
		{"seedream-legacy", "volcengine_ark"},
		{"nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			if got := BackendOf(tt.alias); got != tt.want {
				t.Errorf("BackendOf(%q) = %q, want %q", tt.alias, got, tt.want)
			}
		})
	}
}

func TestModelIDOf(t *testing.T) {
	tests := []struct {
		alias string
		want  string
	}{
		{"seedream", "doubao-seedream-4-5-251128"},
		{"seedream-lite", "doubao-seedream-5.0-lite"},
		{"seedream-legacy", "doubao-seedream-4-0-250828"},
		{"nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			if got := ModelIDOf(tt.alias); got != tt.want {
				t.Errorf("ModelIDOf(%q) = %q, want %q", tt.alias, got, tt.want)
			}
		})
	}
}

func TestMimeTypeByPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"test.jpg", "image/jpeg"},
		{"test.jpeg", "image/jpeg"},
		{"test.png", "image/png"},
		{"test.webp", "image/webp"},
		{"test.unknown", "image/png"},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := MimeTypeByPath(tt.path); got != tt.want {
				t.Errorf("MimeTypeByPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestEncodeImageDataURL(t *testing.T) {
	t.Run("existent", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.png")
		content := []byte{0x89, 0x50, 0x4E, 0x47}
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatal(err)
		}

		dataURL, err := EncodeImageDataURL(path)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.HasPrefix(dataURL, "data:image/png;base64,") {
			t.Errorf("data URL should start with %q, got %q", "data:image/png;base64,", dataURL)
		}

		const expectedB64 = "iVBORw"
		if !strings.HasSuffix(dataURL, expectedB64) {
			t.Errorf("data URL should end with %q, got %q", expectedB64, dataURL)
		}
	})

	t.Run("nonexistent", func(t *testing.T) {
		_, err := EncodeImageDataURL("/nonexistent/path.png")
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})
}

func TestPixelSize(t *testing.T) {
	tests := []struct {
		aspect string
		maxDim int
		sep    string
		want   string
	}{
		{"16:9", 4096, "*", "4096*2304"},
		{"3:4", 2048, "x", "1536x2048"},
		{"", 1024, "*", "1024*1024"},
		{"9:16", 2048, "x", "1152x2048"},
		{"0:16", 128, "x", "512x512"},
	}
	for _, tt := range tests {
		name := tt.aspect
		if name == "" {
			name = "square"
		}
		t.Run(name, func(t *testing.T) {
			if got := PixelSize(tt.aspect, tt.maxDim, tt.sep); got != tt.want {
				t.Errorf("PixelSize(%q, %d, %q) = %q, want %q",
					tt.aspect, tt.maxDim, tt.sep, got, tt.want)
			}
		})
	}
}

func TestVideoBackendOf(t *testing.T) {
	tests := []struct {
		alias string
		want  string
	}{
		{"seedance", "volcengine_ark"},
		{"seedance-fast", "volcengine_ark"},
		{"nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			if got := VideoBackendOf(tt.alias); got != tt.want {
				t.Errorf("VideoBackendOf(%q) = %q, want %q", tt.alias, got, tt.want)
			}
		})
	}
}

func TestVideoModelIDOf(t *testing.T) {
	tests := []struct {
		alias string
		want  string
	}{
		{"seedance", "doubao-seedance-2-0-260128"},
		{"seedance-fast", "doubao-seedance-2-0-fast-260128"},
		{"nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			if got := VideoModelIDOf(tt.alias); got != tt.want {
				t.Errorf("VideoModelIDOf(%q) = %q, want %q", tt.alias, got, tt.want)
			}
		})
	}
}

package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type modelEntry struct {
	Backend string
	ModelID string
}

var ImageModels = map[string]modelEntry{
	"seedream":        {"volcengine_ark", "doubao-seedream-4-5-251128"},
	"seedream-lite":   {"volcengine_ark", "doubao-seedream-5.0-lite"},
	"seedream-legacy": {"volcengine_ark", "doubao-seedream-4-0-250828"},
}

var VideoModels = map[string]modelEntry{
	"seedance":      {"volcengine_ark", "doubao-seedance-2-0-260128"},
	"seedance-fast": {"volcengine_ark", "doubao-seedance-2-0-fast-260128"},
}

var ImageAspectRatios = []string{"1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"}
var VideoAspectRatios = []string{"16:9", "9:16", "4:3", "3:4", "1:1", "21:9", "adaptive"}
var VideoResolutions = []string{"480p", "720p", "1080p", "2K"}

type ImageOptions struct {
	Model          string
	AspectRatio    string
	Resolution     string
	NegativePrompt string
}

type VideoOptions struct {
	Model         string
	AspectRatio   string
	Resolution    string
	Duration      int
	GenerateAudio bool
	Seed          int
}

type TaskInfo struct {
	TaskID   string
	Status   string
	VideoURL string
	Error    string
}

func BackendOf(alias string) string {
	if e, ok := ImageModels[alias]; ok {
		return e.Backend
	}
	return ""
}

func ModelIDOf(alias string) string {
	if e, ok := ImageModels[alias]; ok {
		return e.ModelID
	}
	return ""
}

func VideoBackendOf(alias string) string {
	if e, ok := VideoModels[alias]; ok {
		return e.Backend
	}
	return ""
}

func VideoModelIDOf(alias string) string {
	if e, ok := VideoModels[alias]; ok {
		return e.ModelID
	}
	return ""
}

func PixelSize(aspectRatio string, maxDim int, sep string) string {
	width, height := maxDim, maxDim
	if aspectRatio != "" {
		parts := strings.SplitN(aspectRatio, ":", 2)
		if len(parts) == 2 {
			w, errW := strconv.ParseFloat(parts[0], 64)
			h, errH := strconv.ParseFloat(parts[1], 64)
			if errW == nil && errH == nil && w > 0 && h > 0 {
				if w >= h {
					height = int(math.Round(float64(maxDim) * h / w))
				} else {
					width = int(math.Round(float64(maxDim) * w / h))
				}
			}
		}
	}
	width = round8(width)
	height = round8(height)
	if width < 512 {
		width = 512
	}
	if height < 512 {
		height = 512
	}
	return fmt.Sprintf("%d%s%d", width, sep, height)
}

func round8(v int) int {
	return ((v + 4) / 8) * 8
}

func MimeTypeByPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func EncodeImageDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	mime := MimeTypeByPath(path)
	b64 := base64.RawStdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", mime, b64), nil
}

func httpPostJSON(client *http.Client, url string, body map[string]any, headers map[string]string) (map[string]any, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("HTTP %d (failed to read body: %w)", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func httpGetBytes(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}

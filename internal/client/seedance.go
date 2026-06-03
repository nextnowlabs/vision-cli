package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const seedanceTasksPath = "/contents/generations/tasks"

type SeedanceClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewSeedanceClient(apiKey string) *SeedanceClient {
	return &SeedanceClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 300 * time.Second},
	}
}

func (c *SeedanceClient) Submit(prompt string, inputImages []string, opts VideoOptions) (TaskInfo, error) {
	modelID := VideoModelIDOf(opts.Model)
	content := []map[string]any{}

	if len(inputImages) > 0 {
		for _, p := range inputImages {
			dataURL, err := EncodeImageDataURL(p)
			if err != nil {
				return TaskInfo{}, fmt.Errorf("encode %s: %w", p, err)
			}
			content = append(content, map[string]any{
				"type":      "image_url",
				"image_url": map[string]any{"url": dataURL},
				"role":      "reference_image",
			})
		}
	}
	content = append(content, map[string]any{"type": "text", "text": prompt})

	body := map[string]any{
		"model":          modelID,
		"content":        content,
		"resolution":     opts.Resolution,
		"duration":       opts.Duration,
		"generate_audio": opts.GenerateAudio,
	}
	if opts.AspectRatio != "" {
		body["ratio"] = opts.AspectRatio
	}
	if opts.Seed != 0 {
		body["seed"] = opts.Seed
	}

	payload, err := httpPostJSON(c.httpClient, arkBaseURL+seedanceTasksPath, body, map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.apiKey,
	})
	if err != nil {
		return TaskInfo{}, err
	}

	taskID, _ := payload["id"].(string)
	if taskID == "" {
		taskID, _ = payload["task_id"].(string)
	}
	if taskID == "" {
		return TaskInfo{}, fmt.Errorf("failed to submit Seedance task: %v", payload)
	}
	status, _ := payload["status"].(string)
	if status == "" {
		status = "queued"
	}
	return TaskInfo{TaskID: taskID, Status: status}, nil
}

func (c *SeedanceClient) GetStatus(taskID string) (TaskInfo, error) {
	url := arkBaseURL + seedanceTasksPath + "/" + taskID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return TaskInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TaskInfo{}, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return TaskInfo{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TaskInfo{}, fmt.Errorf("Ark HTTP %d: %s", resp.StatusCode, string(respData))
	}

	var payload map[string]any
	if err := json.Unmarshal(respData, &payload); err != nil {
		return TaskInfo{}, err
	}

	info := TaskInfo{TaskID: taskID}
	if s, ok := payload["status"].(string); ok {
		info.Status = s
	}
	if content, ok := payload["content"].(map[string]any); ok {
		if url, ok := content["video_url"].(string); ok {
			info.VideoURL = url
		}
	}
	if e, ok := payload["error"].(string); ok {
		info.Error = e
	}
	return info, nil
}

func (c *SeedanceClient) Poll(taskID string, intervalSec int, callback func(int, string, int)) (TaskInfo, error) {
	count := 0
	for {
		time.Sleep(time.Duration(intervalSec) * time.Second)
		count++
		status, err := c.GetStatus(taskID)
		if err != nil {
			return TaskInfo{}, err
		}
		if callback != nil {
			callback(count, status.Status, count*intervalSec)
		}
		switch status.Status {
		case "succeeded", "failed", "expired", "cancelled":
			return status, nil
		}
	}
}

func (c *SeedanceClient) DownloadVideo(videoURL, outputPath string) (string, error) {
	resp, err := c.httpClient.Get(videoURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	parent := filepath.Dir(outputPath)
	if parent != "" {
		if err := os.MkdirAll(parent, 0755); err != nil {
			return "", err
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	abs, _ := filepath.Abs(outputPath)
	return abs, nil
}

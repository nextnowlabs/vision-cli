package client

import (
	"fmt"
	"net/http"
	"time"

	imgconv "github.com/nextnowlabs/vision-cli/internal/image"
)

const dashscopeSyncURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"

var dashscopeResCaps = map[string]map[string]int{
	"wan2.7-image-pro": {"1K": 1024, "2K": 2048, "4K": 4096},
	"wan2.7-image":     {"1K": 1024, "2K": 2048, "4K": 2048},
}

type DashScopeClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewDashScopeClient(apiKey string) *DashScopeClient {
	return &DashScopeClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 300 * time.Second},
	}
}

func (c *DashScopeClient) Generate(prompt string, inputImages []string, opts ImageOptions) ([]byte, error) {
	modelID := ModelIDOf(opts.Model)
	if len(inputImages) > 9 {
		return nil, fmt.Errorf("wan2.7 accepts at most 9 reference images")
	}

	caps := dashscopeResCaps[modelID]
	if caps == nil {
		caps = dashscopeResCaps["wan2.7-image"]
	}
	maxDim := caps[opts.Resolution]
	if maxDim == 0 {
		maxDim = 2048
	}
	size := PixelSize(opts.AspectRatio, maxDim, "*")

	content := []map[string]any{}

	if len(inputImages) > 0 {
		for _, p := range inputImages {
			dataURL, err := EncodeImageDataURL(p)
			if err != nil {
				return nil, fmt.Errorf("encode %s: %w", p, err)
			}
			content = append(content, map[string]any{"image": dataURL})
		}
	}
	content = append(content, map[string]any{"text": prompt})

	parameters := map[string]any{
		"size":      size,
		"n":         1,
		"watermark": false,
	}
	if modelID == "wan2.7-image-pro" || modelID == "wan2.7-image" {
		parameters["thinking_mode"] = true
	}
	if opts.NegativePrompt != "" {
		parameters["negative_prompt"] = opts.NegativePrompt
	}

	body := map[string]any{
		"model": modelID,
		"input": map[string]any{
			"messages": []map[string]any{
				{"role": "user", "content": content},
			},
		},
		"parameters": parameters,
	}

	payload, err := httpPostJSON(c.httpClient, dashscopeSyncURL, body, map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.apiKey,
	})
	if err != nil {
		return nil, err
	}

	output, _ := payload["output"].(map[string]any)
	if output == nil {
		return nil, fmt.Errorf("no output in DashScope response: %v", payload)
	}
	choices, _ := output["choices"].([]any)
	if len(choices) == 0 {
		return nil, fmt.Errorf("no choices in DashScope response: %v", payload)
	}
	msg, _ := choices[0].(map[string]any)["message"].(map[string]any)
	if msg == nil {
		return nil, fmt.Errorf("no message in DashScope response: %v", choices[0])
	}
	msgContent, _ := msg["content"].([]any)

	var imgURL string
	for _, part := range msgContent {
		p, ok := part.(map[string]any)
		if !ok {
			continue
		}
		if url, ok := p["image"].(string); ok && url != "" {
			imgURL = url
			break
		}
	}
	if imgURL == "" {
		return nil, fmt.Errorf("no image URL in DashScope response")
	}

	imgData, err := httpGetBytes(c.httpClient, imgURL)
	if err != nil {
		return nil, fmt.Errorf("download image: %w", err)
	}
	return imgconv.ToRGBPNG(imgData)
}

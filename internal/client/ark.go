package client

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	imgconv "github.com/nextnowlabs/vision-cli/internal/image"
)

const arkBaseURL = "https://ark.cn-beijing.volces.com/api/v3"
const arkGenPath = "/images/generations"

var arkResCaps = map[string]map[string]int{
	"doubao-seedream-4-0-250828": {"1K": 1024, "2K": 2048, "4K": 4096},
	"doubao-seedream-4-5-251128": {"1K": 1024, "2K": 2048, "4K": 4096},
	"doubao-seedream-5.0-lite":   {"1K": 1024, "2K": 2048, "4K": 4096},
}

const arkMaxRefs = 14

type VolcengineArkClient struct {
	apiKey     string
	endpointID string
	httpClient *http.Client
}

func NewVolcengineArkClient(apiKey, endpointID string) *VolcengineArkClient {
	return &VolcengineArkClient{
		apiKey:     apiKey,
		endpointID: endpointID,
		httpClient: &http.Client{Timeout: 300 * time.Second},
	}
}

func (c *VolcengineArkClient) Generate(prompt string, inputImages []string, opts ImageOptions) ([]byte, error) {
	modelID := ModelIDOf(opts.Model)
	if len(inputImages) > arkMaxRefs {
		return nil, fmt.Errorf("Seedream accepts at most %d reference images", arkMaxRefs)
	}

	caps := arkResCaps[modelID]
	if caps == nil {
		caps = arkResCaps["doubao-seedream-4-5-251128"]
	}
	maxDim := caps[opts.Resolution]
	if maxDim == 0 {
		maxDim = 2048
	}
	size := PixelSize(opts.AspectRatio, maxDim, "x")

	fullPrompt := prompt
	if opts.NegativePrompt != "" {
		fullPrompt = fmt.Sprintf("%s\n\nNegative: %s", prompt, opts.NegativePrompt)
	}

	modelName := c.endpointID
	if modelName == "" {
		modelName = modelID
	}

	body := map[string]any{
		"model":                        modelName,
		"prompt":                       fullPrompt,
		"size":                        size,
		"response_format":             "url",
		"watermark":                   false,
		"sequential_image_generation": "disabled",
	}

	if len(inputImages) > 0 {
		dataURLs := make([]string, len(inputImages))
		for i, p := range inputImages {
			url, err := EncodeImageDataURL(p)
			if err != nil {
				return nil, fmt.Errorf("encode %s: %w", p, err)
			}
			dataURLs[i] = url
		}
		if len(dataURLs) == 1 {
			body["image"] = dataURLs[0]
		} else {
			body["image"] = dataURLs
		}
	}

	payload, err := httpPostJSON(c.httpClient, arkBaseURL+arkGenPath, body, map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.apiKey,
	})
	if err != nil {
		return nil, err
	}

	data, _ := payload["data"].([]any)
	if len(data) == 0 {
		return nil, fmt.Errorf("no image in Ark response: %v", payload)
	}
	first, _ := data[0].(map[string]any)
	if url, ok := first["url"].(string); ok {
		imgData, err := httpGetBytes(c.httpClient, url)
		if err != nil {
			return nil, fmt.Errorf("download image: %w", err)
		}
		return imgconv.ToRGBPNG(imgData)
	}
	if b64, ok := first["b64_json"].(string); ok {
		imgData, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("decode b64_json: %w", err)
		}
		return imgconv.ToRGBPNG(imgData)
	}
	return nil, fmt.Errorf("unrecognized Ark response: %v", first)
}

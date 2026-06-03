package client

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const ttsBaseURL = "https://openspeech.bytedance.com/api/v3/tts/unidirectional/sse"

type TTSOptions struct {
	VoiceType    string
	Encoding     string
	Rate         int
	SpeedRatio   float64
	VolumeRatio  float64
	Language     string
	Emotion      string
	EmotionScale float64
	Pitch        int
}

type TTSClient struct {
	apiKey     string
	resourceID string
	http       *http.Client
}

func NewTTSClient(apiKey, resourceID string) *TTSClient {
	return &TTSClient{
		apiKey:     apiKey,
		resourceID: resourceID,
		http:       &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *TTSClient) Synthesize(text string, opts TTSOptions) ([]byte, string, error) {
	if opts.Encoding == "" {
		opts.Encoding = "mp3"
	}
	if opts.Rate == 0 {
		opts.Rate = 24000
	}
	if opts.SpeedRatio == 0 {
		opts.SpeedRatio = 1.0
	}
	if opts.VolumeRatio == 0 {
		opts.VolumeRatio = 1.0
	}

	audioParams := map[string]any{
		"format":      opts.Encoding,
		"sample_rate": opts.Rate,
		"speech_rate": speedRatioToRate(opts.SpeedRatio),
		"loudness_rate": speedRatioToRate(opts.VolumeRatio),
	}
	if opts.Emotion != "" {
		audioParams["emotion"] = opts.Emotion
	}
	if opts.EmotionScale > 0 {
		audioParams["emotion_scale"] = opts.EmotionScale
	}

	additions := map[string]any{}
	if opts.Language != "" {
		additions["explicit_language"] = opts.Language
	}
	if opts.Pitch != 0 {
		additions["post_process"] = map[string]any{"pitch": opts.Pitch}
	}

	reqParams := map[string]any{
		"text":         text,
		"speaker":      opts.VoiceType,
		"audio_params": audioParams,
	}
	if len(additions) > 0 {
		a, _ := json.Marshal(additions)
		reqParams["additions"] = string(a)
	}

	body := map[string]any{
		"user": map[string]any{
			"uid": "vg-cli",
		},
		"namespace":    "BidirectionalTTS",
		"req_params":   reqParams,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequest("POST", ttsBaseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("X-Api-Resource-Id", c.resourceID)
	req.Header.Set("X-Api-Request-Id", reqID())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var audioChunks []byte
	var duration string
	var lastErr error

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "event:") {
			continue
		}
		data, ok := strings.CutPrefix(line, "data:")
		if !ok {
			continue
		}

		var event map[string]any
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		code, _ := event["code"].(float64)
		if code != 0 && code != 20000000 {
			msg, _ := event["message"].(string)
			lastErr = fmt.Errorf("TTS error %v: %s", code, msg)
			continue
		}

		if b64, ok := event["data"].(string); ok && b64 != "" {
			chunk, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				lastErr = fmt.Errorf("decode audio chunk: %w", err)
				continue
			}
			audioChunks = append(audioChunks, chunk...)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("read SSE stream: %w", err)
	}

	if lastErr != nil {
		return nil, "", lastErr
	}

	if len(audioChunks) == 0 {
		return nil, "", fmt.Errorf("no audio data received from TTS")
	}

	return audioChunks, duration, nil
}

func speedRatioToRate(ratio float64) int {
	return int(math.Round((ratio - 1.0) * 100))
}

func reqID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

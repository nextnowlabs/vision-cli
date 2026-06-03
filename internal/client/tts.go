package client

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const ttsBaseURL = "https://openspeech.bytedance.com/api/v1/tts"

type TTSOptions struct {
	VoiceType   string
	Encoding    string
	Rate        int
	SpeedRatio  float64
	VolumeRatio float64
	PitchRatio  float64
	Language    string
	Emotion     string
}

type TTSClient struct {
	appID   string
	token   string
	cluster string
	http    *http.Client
}

func NewTTSClient(appID, token, cluster string) *TTSClient {
	return &TTSClient{
		appID:   appID,
		token:   token,
		cluster: cluster,
		http:    &http.Client{Timeout: 60 * time.Second},
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
	if opts.PitchRatio == 0 {
		opts.PitchRatio = 1.0
	}

	audio := map[string]any{
		"voice_type":   opts.VoiceType,
		"encoding":     opts.Encoding,
		"rate":         opts.Rate,
		"speed_ratio":  opts.SpeedRatio,
		"volume_ratio": opts.VolumeRatio,
		"pitch_ratio":  opts.PitchRatio,
	}
	if opts.Language != "" {
		audio["language"] = opts.Language
	}
	if opts.Emotion != "" {
		audio["emotion"] = opts.Emotion
	}

	body := map[string]any{
		"app": map[string]any{
			"appid":   c.appID,
			"token":   c.token,
			"cluster": c.cluster,
		},
		"user": map[string]any{
			"uid": "vg-cli",
		},
		"audio": audio,
		"request": map[string]any{
			"reqid":    reqID(),
			"text":     text,
			"text_type": "plain",
			"operation": "query",
		},
	}

	payload, err := httpPostJSON(c.http, ttsBaseURL, body, map[string]string{
		"Authorization": "Bearer;" + c.token,
	})
	if err != nil {
		return nil, "", err
	}

	code, _ := payload["code"].(float64)
	if code != 3000 {
		msg, _ := payload["message"].(string)
		return nil, "", fmt.Errorf("TTS error %v: %s", code, msg)
	}

	b64, _ := payload["data"].(string)
	if b64 == "" {
		return nil, "", fmt.Errorf("no audio data in TTS response")
	}

	audioData, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, "", fmt.Errorf("decode audio: %w", err)
	}

	duration := ""
	if addition, ok := payload["addition"].(map[string]any); ok {
		if d, ok := addition["duration"].(string); ok {
			duration = d
		}
	}

	return audioData, duration, nil
}

func reqID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type VoiceInfo struct {
	SpeakerID string `json:"speaker_id"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	Gender    string `json:"gender"`
}

type VoiceListClient struct {
	ak     string
	sk     string
	appID  string
	http   *http.Client
}

func NewVoiceListClient(ak, sk, appID string) *VoiceListClient {
	return &VoiceListClient{
		ak:    ak,
		sk:    sk,
		appID: appID,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

const volcAPIBase = "https://open.volcengineapi.com"

func (c *VoiceListClient) ListSpeakers() ([]VoiceInfo, error) {
	query := map[string]string{
		"Action":  "ListBigModelTTSTimbres",
		"Version": "2023-11-07",
	}

	body := map[string]any{
		"AppId": c.appID,
	}

	payload, err := httpPostJSONVolc(c.http, volcAPIBase, "ListBigModelTTSTimbres", "2023-11-07",
		"cn-north-1", "speech_saas_prod", query, body, c.ak, c.sk)
	if err != nil {
		return nil, fmt.Errorf("list speakers: %w", err)
	}

	result, ok := payload["Result"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response: no Result field")
	}

	speakers, ok := result["Speakers"].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response: Speakers field missing or not array")
	}

	var voices []VoiceInfo
	for _, s := range speakers {
		item, ok := s.(map[string]any)
		if !ok {
			continue
		}
		v := VoiceInfo{}
		if sid, ok := item["SpeakerId"].(string); ok {
			v.SpeakerID = sid
		}
		if name, ok := item["Speaker"].(string); ok {
			v.Name = name
		}
		if lang, ok := item["Language"].(string); ok {
			v.Language = lang
		}
		if gender, ok := item["Gender"].(string); ok {
			v.Gender = gender
		}
		voices = append(voices, v)
	}

	return voices, nil
}

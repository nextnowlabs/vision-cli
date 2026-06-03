package cli

import (
	"os"
	"strings"

	"github.com/nextnowlabs/vision-cli/internal/client"
)

var apiKeyHints = map[string]string{
	"dashscope":     "No DashScope API key. Set via: vg config set dashscope_api_key <key>\nOr export DASHSCOPE_API_KEY=<key>",
	"volcengine_ark": "No Volcengine Ark API key. Set via: vg config set ark_api_key <key>\nOr export ARK_API_KEY=<key>\nNote: 还需在火山方舟控制台「开通管理」中开通对应服务",
}

func resolvePrompt(prompt string) string {
	if strings.HasPrefix(prompt, "@") {
		data, err := os.ReadFile(prompt[1:])
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return prompt
}

func getCfgStr(cfg map[string]any, key, defaultVal string) string {
	if v, ok := cfg[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

type ImageGenerator interface {
	Generate(prompt string, inputImages []string, opts client.ImageOptions) ([]byte, error)
}

func newImageClient(backend, apiKey string, cfg map[string]any) ImageGenerator {
	switch backend {
	case "dashscope":
		return client.NewDashScopeClient(apiKey)
	case "volcengine_ark":
		endpointID := getCfgStr(cfg, "ark_endpoint_id", "")
		return client.NewVolcengineArkClient(apiKey, endpointID)
	default:
		panic("unknown backend: " + backend)
	}
}

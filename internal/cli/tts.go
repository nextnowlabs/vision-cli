package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nextnowlabs/vision-cli/internal/client"
	"github.com/spf13/cobra"
)

var ttsCmd = &cobra.Command{
	Use:   "tts",
	Short: "Text-to-speech synthesis (Volcengine TTS)",
}

var ttsGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Synthesize speech from text",
	RunE:  runTTSGen,
}

var ttsVoicesCmd = &cobra.Command{
	Use:   "voices",
	Short: "List available voice types",
	RunE:  runTTSVoices,
}

var (
	ttsPrompt    string
	ttsOutput    string
	ttsVoiceType string
	ttsEncoding  string
	ttsRate      int
	ttsSpeed     float64
	ttsVolume    float64
	ttsPitch     float64
	ttsLanguage  string
	ttsEmotion   string
)

func init() {
	ttsGenCmd.Flags().StringVarP(&ttsPrompt, "prompt", "p", "", "Text to synthesize or @file.txt")
	ttsGenCmd.MarkFlagRequired("prompt")
	ttsGenCmd.Flags().StringVarP(&ttsOutput, "output", "o", "", "Output audio path (default: auto timestamp)")
	ttsGenCmd.Flags().StringVar(&ttsVoiceType, "voice-type", "BV700_streaming", "Voice type (speaker ID)")
	ttsGenCmd.Flags().StringVar(&ttsEncoding, "encoding", "mp3", "Audio encoding (wav/pcm/ogg_opus/mp3)")
	ttsGenCmd.Flags().IntVar(&ttsRate, "rate", 24000, "Audio sample rate (8000/16000/24000)")
	ttsGenCmd.Flags().Float64Var(&ttsSpeed, "speed", 1.0, "Speech speed ratio (0.2-3.0)")
	ttsGenCmd.Flags().Float64Var(&ttsVolume, "volume", 1.0, "Speech volume ratio (0.1-3.0)")
	ttsGenCmd.Flags().Float64Var(&ttsPitch, "pitch", 1.0, "Speech pitch ratio (0.1-3.0)")
	ttsGenCmd.Flags().StringVar(&ttsLanguage, "language", "", "Language code (e.g. cn, en)")
	ttsGenCmd.Flags().StringVar(&ttsEmotion, "emotion", "", "Emotion/style (e.g. happy, sad)")

	ttsCmd.AddCommand(ttsGenCmd)
	ttsCmd.AddCommand(ttsVoicesCmd)
}

var (
	_ = ttsCmd
)

func runTTSGen(cmd *cobra.Command, args []string) error {
	cfg, _ := appStore.LoadConfig()

	appID := getCfgStr(cfg, "tts_appid", "")
	cluster := getCfgStr(cfg, "tts_cluster", "")
	token := getCfgStr(cfg, "tts_token", "")
	if token == "" {
		if v := os.Getenv("TTS_TOKEN"); v != "" {
			token = v
		}
	}
	if appID == "" || cluster == "" || token == "" {
		cmd.PrintErrln("TTS requires app_id, cluster, and token. Set via:\n  vg config set tts_appid <id>\n  vg config set tts_cluster <cluster>\n  vg config set tts_token <token>\nOr export TTS_TOKEN=<token>")
		os.Exit(1)
	}

	text := resolvePrompt(ttsPrompt)
	outputDir := getCfgStr(cfg, "output_dir", ".")

	output := ttsOutput
	if output == "" {
		ts := time.Now().Format("20060102_150405")
		ext := fileExtension(ttsEncoding)
		output = filepath.Join(outputDir, fmt.Sprintf("vg_tts_%s.%s", ts, ext))
	}
	parent := filepath.Dir(output)
	if parent != "" {
		os.MkdirAll(parent, 0755)
	}

	c := client.NewTTSClient(appID, token, cluster)

	cmd.Printf("Synthesizing... (voice=%s, encoding=%s, rate=%d)\n", ttsVoiceType, ttsEncoding, ttsRate)

	audioData, duration, err := c.Synthesize(text, client.TTSOptions{
		VoiceType:   ttsVoiceType,
		Encoding:    ttsEncoding,
		Rate:        ttsRate,
		SpeedRatio:  ttsSpeed,
		VolumeRatio: ttsVolume,
		PitchRatio:  ttsPitch,
		Language:    ttsLanguage,
		Emotion:     ttsEmotion,
	})
	if err != nil {
		cmd.PrintErrf("Synthesis failed: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(output, audioData, 0644); err != nil {
		return err
	}

	abs, _ := filepath.Abs(output)
	durStr := duration
	if durStr != "" {
		durStr = fmt.Sprintf(" (%sms)", durStr)
	}
	cmd.Printf("  %s%s\n", abs, durStr)
	return nil
}

func runTTSVoices(cmd *cobra.Command, args []string) error {
	cfg, _ := appStore.LoadConfig()

	appID := getCfgStr(cfg, "tts_appid", "")
	ak := getCfgStr(cfg, "tts_access_key", "")
	sk := getCfgStr(cfg, "tts_secret_key", "")
	if appID == "" || ak == "" || sk == "" {
		cmd.PrintErrln("Voice list requires app_id, access_key, and secret_key.\nSet via:\n  vg config set tts_appid <id>\n  vg config set tts_access_key <access_key>\n  vg config set tts_secret_key <secret_key>")
		os.Exit(1)
	}

	c := client.NewVoiceListClient(ak, sk, appID)

	cmd.Println("Fetching voice list...")
	voices, err := c.ListSpeakers()
	if err != nil {
		cmd.PrintErrf("Failed: %v\n", err)
		os.Exit(1)
	}

	if len(voices) == 0 {
		cmd.Println("No voices found.")
		return nil
	}

	for _, v := range voices {
		gender := v.Gender
		if gender == "" {
			gender = "-"
		}
		lang := v.Language
		if lang == "" {
			lang = "-"
		}
		cmd.Printf("  %-30s  %-20s  %-6s  %-6s\n", v.SpeakerID, v.Name, gender, lang)
	}
	cmd.Printf("\nTotal: %d voices\n", len(voices))
	return nil
}

func fileExtension(encoding string) string {
	switch encoding {
	case "wav":
		return "wav"
	case "pcm":
		return "pcm"
	case "ogg_opus":
		return "opus"
	default:
		return "mp3"
	}
}

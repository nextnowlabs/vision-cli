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
	Short: "Text-to-speech synthesis (Volcengine TTS V3)",
}

var ttsGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Synthesize speech from text",
	RunE:  runTTSGen,
}

var (
	ttsPrompt       string
	ttsOutput       string
	ttsVoiceType    string
	ttsEncoding     string
	ttsRate         int
	ttsSpeed        float64
	ttsVolume       float64
	ttsLanguage     string
	ttsEmotion      string
	ttsEmotionScale float64
	ttsPitch        int
)

func init() {
	ttsGenCmd.Flags().StringVarP(&ttsPrompt, "prompt", "p", "", "Text to synthesize or @file.txt")
	ttsGenCmd.MarkFlagRequired("prompt")
	ttsGenCmd.Flags().StringVarP(&ttsOutput, "output", "o", "", "Output audio path (default: auto timestamp)")
	ttsGenCmd.Flags().StringVar(&ttsVoiceType, "voice-type", "zh_female_shuangkuaisisi_uranus_bigtts", "Voice type (speaker ID)")
	ttsGenCmd.Flags().StringVar(&ttsEncoding, "encoding", "mp3", "Audio encoding (mp3/ogg_opus/pcm)")
	ttsGenCmd.Flags().IntVar(&ttsRate, "rate", 24000, "Audio sample rate (8000/16000/22050/24000/32000/44100/48000)")
	ttsGenCmd.Flags().Float64Var(&ttsSpeed, "speed", 1.0, "Speech speed ratio (0.2-3.0)")
	ttsGenCmd.Flags().Float64Var(&ttsVolume, "volume", 1.0, "Speech volume ratio (0.1-3.0)")
	ttsGenCmd.Flags().StringVar(&ttsLanguage, "language", "", "Language code (e.g. zh-cn, en, ja)")
	ttsGenCmd.Flags().StringVar(&ttsEmotion, "emotion", "", "Emotion/style (e.g. happy, sad)")
	ttsGenCmd.Flags().Float64Var(&ttsEmotionScale, "emotion-scale", 0, "Emotion intensity 1-5 (default 4 when emotion is set)")
	ttsGenCmd.Flags().IntVar(&ttsPitch, "pitch", 0, "Pitch shift in semitones (-12 to 12)")

	ttsCmd.AddCommand(ttsGenCmd)
}

var (
	_ = ttsCmd
)

func runTTSGen(cmd *cobra.Command, args []string) error {
	cfg, _ := appStore.LoadConfig()

	apiKey := getCfgStr(cfg, "tts_api_key", "")
	if apiKey == "" {
		if v := os.Getenv("TTS_API_KEY"); v != "" {
			apiKey = v
		}
	}
	resourceID := getCfgStr(cfg, "tts_resource_id", "seed-tts-2.0")
	if apiKey == "" {
		cmd.PrintErrln("TTS requires api_key. Set via:\n  vg config set tts_api_key <key>\n  vg config set tts_resource_id <id>  (default: seed-tts-2.0)\nOr export TTS_API_KEY=<key>")
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

	c := client.NewTTSClient(apiKey, resourceID)

	cmd.Printf("Synthesizing... (voice=%s, encoding=%s, rate=%d)\n", ttsVoiceType, ttsEncoding, ttsRate)

	audioData, duration, err := c.Synthesize(text, client.TTSOptions{
		VoiceType:    ttsVoiceType,
		Encoding:     ttsEncoding,
		Rate:         ttsRate,
		SpeedRatio:   ttsSpeed,
		VolumeRatio:  ttsVolume,
		Language:     ttsLanguage,
		Emotion:      ttsEmotion,
		EmotionScale: ttsEmotionScale,
		Pitch:        ttsPitch,
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

func fileExtension(encoding string) string {
	switch encoding {
	case "pcm":
		return "pcm"
	case "ogg_opus":
		return "opus"
	default:
		return "mp3"
	}
}

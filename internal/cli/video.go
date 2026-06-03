package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/nextnowlabs/vision-cli/internal/client"
	"github.com/nextnowlabs/vision-cli/internal/store"
	"github.com/spf13/cobra"
)

var videoCmd = &cobra.Command{
	Use:   "video",
	Short: "Video generation (Seedance 2.0 on Volcengine Ark)",
}

var (
	videoGenPrompt string
	videoGenOutput string
	videoGenImages []string
	videoGenAR     string
	videoGenRes    string
	videoGenDur    int
	videoGenAudio  bool
	videoGenSeed   int
	videoGenModel  string
	videoGenPoll   bool
)

var videoGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a video from a text/image prompt",
	RunE:  runVideoGen,
}

func init() {
	videoGenCmd.Flags().StringVarP(&videoGenPrompt, "prompt", "p", "", "Prompt text or @file.txt")
	videoGenCmd.MarkFlagRequired("prompt")
	videoGenCmd.Flags().StringVarP(&videoGenOutput, "output", "o", "", "Output MP4 path")
	videoGenCmd.Flags().StringSliceVarP(&videoGenImages, "input", "i", nil, "Reference image")
	videoGenCmd.Flags().StringVar(&videoGenAR, "ar", "", "Aspect ratio")
	videoGenCmd.Flags().StringVar(&videoGenRes, "res", "720p", "Video resolution")
	videoGenCmd.Flags().IntVar(&videoGenDur, "duration", 5, "Duration in seconds (4-15)")
	videoGenCmd.Flags().BoolVar(&videoGenAudio, "audio", false, "Generate native synced audio")
	videoGenCmd.Flags().IntVar(&videoGenSeed, "seed", 0, "Random seed")
	videoGenCmd.Flags().StringVar(&videoGenModel, "model", "seedance", "Video model alias")
	videoGenCmd.Flags().BoolVar(&videoGenPoll, "poll", true, "Poll and auto-download")

	videoCmd.AddCommand(videoGenCmd)
	videoCmd.AddCommand(videoStatusCmd)
	videoCmd.AddCommand(videoDownloadCmd)
}

func runVideoGen(cmd *cobra.Command, args []string) error {
	cfg, _ := appStore.LoadConfig()
	outputDir := getCfgStr(cfg, "output_dir", ".")

	prompt := resolvePrompt(videoGenPrompt)
	apiKey, ok := appStore.GetAPIKey("volcengine_ark")
	if !ok {
		cmd.PrintErrln(apiKeyHints["volcengine_ark"])
		os.Exit(1)
	}

	c := client.NewSeedanceClient(apiKey)

	output := videoGenOutput
	if output == "" {
		ts := time.Now().Format("20060102_150405")
		output = filepath.Join(outputDir, fmt.Sprintf("vg_video_%s.mp4", ts))
	}
	parent := filepath.Dir(output)
	if parent != "" {
		os.MkdirAll(parent, 0755)
	}

	arDisplay := videoGenAR
	if arDisplay == "" {
		arDisplay = "auto"
	}
	cmd.Printf("Submitting (model=%s, ratio=%s, res=%s, duration=%ds, audio=%v)...\n",
		videoGenModel, arDisplay, videoGenRes, videoGenDur, videoGenAudio)

	info, err := c.Submit(prompt, videoGenImages, client.VideoOptions{
		Model:         videoGenModel,
		AspectRatio:   videoGenAR,
		Resolution:    videoGenRes,
		Duration:      videoGenDur,
		GenerateAudio: videoGenAudio,
		Seed:          videoGenSeed,
	})
	if err != nil {
		cmd.PrintErrf("Submit failed: %v\n", err)
		os.Exit(1)
	}

	cmd.Printf("Task: %s  Status: %s\n", info.TaskID, info.Status)
	appStore.AddRecord(store.Record{
		Prompt:      prompt,
		InputImages: videoGenImages,
		Mode:        "video",
		Status:      "submitted",
		Model:       client.VideoModelIDOf(videoGenModel),
		Backend:     "volcengine_ark",
		VideoTaskID: info.TaskID,
		Duration:    videoGenDur,
		Resolution:  videoGenRes,
	})

	if !videoGenPoll {
		cmd.Printf("Use: vg video status %s\n", info.TaskID)
		cmd.Printf("     vg video download %s -o %s\n", info.TaskID, output)
		return nil
	}

	interval := 15
	if v, ok := cfg["poll_interval"]; ok {
		switch val := v.(type) {
		case float64:
			interval = int(val)
		case int:
			interval = val
		}
	}
	cmd.Printf("Polling every %ds...\n", interval)

	final, err := c.Poll(info.TaskID, interval, func(count int, state string, elapsed int) {
		cmd.Printf("  [%ds] %s\n", elapsed, state)
	})
	if err != nil {
		return err
	}
	if final.Status != "succeeded" {
		cmd.PrintErrf("Failed: %s  %s\n", final.Status, final.Error)
		os.Exit(1)
	}
	if final.VideoURL == "" {
		cmd.PrintErrln("Succeeded but no video_url")
		os.Exit(1)
	}

	cmd.Printf("Downloading %s\n", final.VideoURL)
	path, err := c.DownloadVideo(final.VideoURL, output)
	if err != nil {
		return err
	}
	cmd.Printf("  %s\n", path)

	appStore.AddRecord(store.Record{
		Prompt:      fmt.Sprintf("[video:downloaded] %s", info.TaskID),
		InputImages: nil,
		Mode:        "video",
		Status:      "success",
		Model:       client.VideoModelIDOf(videoGenModel),
		Backend:     "volcengine_ark",
		VideoTaskID: info.TaskID,
	})
	return nil
}

var videoStatusCmd = &cobra.Command{
	Use:   "status <task_id>",
	Short: "Check a Seedance video task's status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, ok := appStore.GetAPIKey("volcengine_ark")
		if !ok {
			cmd.PrintErrln(apiKeyHints["volcengine_ark"])
			os.Exit(1)
		}
		c := client.NewSeedanceClient(apiKey)
		info, err := c.GetStatus(args[0])
		if err != nil {
			return err
		}
		cmd.Printf("  task_id: %s\n", info.TaskID)
		cmd.Printf("  status: %s\n", info.Status)
		if info.VideoURL != "" {
			cmd.Printf("  video_url: %s\n", info.VideoURL)
		}
		if info.Error != "" {
			cmd.Printf("  error: %s\n", info.Error)
		}
		return nil
	},
}

var videoDownloadCmd = &cobra.Command{
	Use:   "download <task_id>",
	Short: "Download a finished Seedance video",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := appStore.LoadConfig()
		outputDir := getCfgStr(cfg, "output_dir", ".")
		apiKey, ok := appStore.GetAPIKey("volcengine_ark")
		if !ok {
			cmd.PrintErrln(apiKeyHints["volcengine_ark"])
			os.Exit(1)
		}
		c := client.NewSeedanceClient(apiKey)
		info, err := c.GetStatus(args[0])
		if err != nil {
			return err
		}
		if info.Status != "succeeded" {
			cmd.PrintErrf("Task not ready: status=%s\n", info.Status)
			if info.Error != "" {
				cmd.PrintErrf("Error: %s\n", info.Error)
			}
			os.Exit(1)
		}
		if info.VideoURL == "" {
			cmd.PrintErrln("No video URL in task")
			os.Exit(1)
		}
		output := videoGenOutput
		if output == "" {
			output = filepath.Join(outputDir, args[0]+".mp4")
		}
		parent := filepath.Dir(output)
		if parent != "" {
			os.MkdirAll(parent, 0755)
		}
		path, err := c.DownloadVideo(info.VideoURL, output)
		if err != nil {
			return err
		}
		cmd.Printf("  %s\n", path)
		return nil
	},
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nextnowlabs/vision-cli/internal/client"
	"github.com/nextnowlabs/vision-cli/internal/store"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate image(s) from a text prompt",
	RunE:  runGen,
}

var (
	genPrompt      string
	genOutput      string
	genInputImages []string
	genAspectRatio string
	genResolution  string
	genRepeat      int
	genModel       string
	genNegative    string
)

func init() {
	genCmd.Flags().StringVarP(&genPrompt, "prompt", "p", "", "Prompt text or @file.txt")
	genCmd.MarkFlagRequired("prompt")
	genCmd.Flags().StringVarP(&genOutput, "output", "o", "", "Output path (default: auto timestamp)")
	genCmd.Flags().StringSliceVarP(&genInputImages, "input", "i", nil, "Reference image (repeatable)")
	genCmd.Flags().StringVar(&genAspectRatio, "ar", "", "Aspect ratio")
	genCmd.Flags().StringVar(&genResolution, "res", "", "Output resolution (default from config or 1K)")
	genCmd.Flags().IntVar(&genRepeat, "repeat", 1, "Generate N images (1-8)")
	genCmd.Flags().StringVar(&genModel, "model", "", "Model to use")
	genCmd.Flags().StringVar(&genNegative, "negative-prompt", "", "Negative prompt")
}

func runGen(cmd *cobra.Command, args []string) error {
	cfg, _ := appStore.LoadConfig()

	prompt := resolvePrompt(genPrompt)
	modelName := genModel
	if modelName == "" {
		modelName = getCfgStr(cfg, "default_model", "seedream")
	}
	aspectRatio := genAspectRatio
	if aspectRatio == "" {
		aspectRatio = getCfgStr(cfg, "default_ar", "")
	}
	resolution := genResolution
	if resolution == "" {
		resolution = getCfgStr(cfg, "default_res", "2K")
	}
	outputDir := getCfgStr(cfg, "output_dir", ".")

	backend := client.BackendOf(modelName)
	apiKey, ok := appStore.GetAPIKey(backend)
	if !ok {
		cmd.PrintErrln(apiKeyHints[backend])
		os.Exit(1)
	}

	maxRefs := 9
	if backend == "volcengine_ark" {
		maxRefs = 14
	}
	if len(genInputImages) > maxRefs {
		cmd.PrintErrf("Max %d reference images for %s\n", maxRefs, modelName)
		os.Exit(1)
	}

	c := newImageClient(backend, apiKey, cfg)

	output := genOutput
	if output == "" {
		ts := time.Now().Format("20060102_150405")
		output = filepath.Join(outputDir, fmt.Sprintf("vg_%s.png", ts))
	}
	ext := filepath.Ext(output)
	base := strings.TrimSuffix(output, ext)
	if ext == "" {
		ext = ".png"
	}
	parent := filepath.Dir(output)
	if parent != "" {
		os.MkdirAll(parent, 0755)
	}

	doGenerate := func(index int) (int, string) {
		var path string
		if genRepeat > 1 {
			path = fmt.Sprintf("%s_%d%s", base, index, ext)
		} else {
			path = base + ext
		}
		imgBytes, err := c.Generate(prompt, genInputImages, client.ImageOptions{
			Model:          modelName,
			AspectRatio:    aspectRatio,
			Resolution:     resolution,
			NegativePrompt: genNegative,
		})
		if err != nil {
			return index, ""
		}
		if err := os.WriteFile(path, imgBytes, 0644); err != nil {
			return index, ""
		}
		return index, path
	}

	cmd.Printf("Generating... (model=%s, res=%s, ar=%s, repeat=%d)\n",
		modelName, resolution, aspectRatio, genRepeat)

	var outputFiles []string
	var errors int
	var mu sync.Mutex

	if genRepeat == 1 {
		_, path := doGenerate(1)
		if path != "" {
			abs, _ := filepath.Abs(path)
			cmd.Printf("  %s\n", abs)
			outputFiles = append(outputFiles, path)
		} else {
			errors = 1
		}
	} else {
		workers := genRepeat
		if workers > 4 {
			workers = 4
		}
		sem := make(chan struct{}, workers)
		var wg sync.WaitGroup

		for i := 1; i <= genRepeat; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				_, p := doGenerate(idx)
				mu.Lock()
				if p != "" {
					abs, _ := filepath.Abs(p)
					cmd.Printf("  [%d/%d] %s\n", idx, genRepeat, abs)
					outputFiles = append(outputFiles, p)
				} else {
					cmd.Printf("  [%d/%d] Failed\n", idx, genRepeat)
					errors++
				}
				mu.Unlock()
			}(i)
		}
		wg.Wait()
	}

	status := "success"
	if len(outputFiles) == 0 {
		status = "failed"
	}
	appStore.AddRecord(store.Record{
		Prompt:      prompt,
		InputImages: genInputImages,
		AspectRatio: aspectRatio,
		Resolution:  resolution,
		Mode:        "direct",
		Status:      status,
		Model:       client.ModelIDOf(modelName),
		Backend:     backend,
	})

	cmd.Printf("\nDone: %d ok, %d failed\n", len(outputFiles), errors)
	return nil
}

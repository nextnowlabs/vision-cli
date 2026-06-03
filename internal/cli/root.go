package cli

import (
	"os"

	"github.com/nextnowlabs/vision-cli/internal/store"
	"github.com/spf13/cobra"
)

var (
	appStore *store.Store
	Version  = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "vg",
	Short:   "Multi-backend image, video & TTS generation CLI",
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.DefaultStore()
		if err != nil {
			return err
		}
		appStore = s
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(videoCmd)
	rootCmd.AddCommand(ttsCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(statsCmd)
}

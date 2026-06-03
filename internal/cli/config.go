package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var validConfigKeys = map[string]bool{
	"dashscope_api_key":       true,
	"ark_api_key":             true,
	"ark_endpoint_id":         true,
	"tts_appid":               true,
	"tts_cluster":             true,
	"tts_token":               true,
	"tts_access_key":          true,
	"tts_secret_key":          true,
	"default_ar":              true,
	"default_res":             true,
	"default_model":           true,
	"poll_interval":           true,
	"output_dir":              true,
}

var secretConfigKeys = map[string]bool{
	"dashscope_api_key":      true,
	"ark_api_key":            true,
	"tts_token":              true,
	"tts_secret_key":         true,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := appStore.LoadConfig()
		if err != nil {
			return err
		}
		if len(cfg) == 0 {
			cmd.Println("No config set. Use: vg config set <key> <value>")
			return nil
		}
		keys := make([]string, 0, len(cfg))
		for k := range cfg {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := cfg[k]
			display := fmt.Sprintf("%v", v)
			if secretConfigKeys[k] {
				display = "***"
			}
			cmd.Printf("  %s: %s\n", k, display)
		}
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if !validConfigKeys[key] {
			cmd.Printf("Unknown key: %s\n", key)
			keys := make([]string, 0, len(validConfigKeys))
			for k := range validConfigKeys {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			cmd.Printf("Valid: %s\n", strings.Join(keys, ", "))
			return fmt.Errorf("invalid key")
		}
		if key == "poll_interval" {
			v, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("poll_interval must be an integer: %w", err)
			}
			if err := appStore.SetConfig(key, v); err != nil {
				return err
			}
		} else {
			if err := appStore.SetConfig(key, value); err != nil {
				return err
			}
		}
		display := value
		if secretConfigKeys[key] {
			display = "***"
		}
		cmd.Printf("%s = %s\n", key, display)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}

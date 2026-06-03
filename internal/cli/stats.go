package cli

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show usage statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := appStore.GetStats()
		if err != nil {
			return err
		}
		if s.TotalCalls == 0 {
			cmd.Println("No data yet")
			return nil
		}
		cmd.Printf("Total calls:  %d\n", s.TotalCalls)
		cmd.Printf("  Success:    %d\n", s.Success)
		cmd.Printf("  Failed:     %d\n", s.Failed)
		cmd.Printf("  Direct:     %d\n", s.Direct)
		cmd.Printf("  Batch:      %d\n", s.Batch)
		cmd.Printf("  Video:      %d\n", s.Video)
		cmd.Printf("  Images:     %d\n", s.TotalImages)

		if len(s.Monthly) > 0 {
			cmd.Println("\nMonthly:")
			months := sortedKeys(s.Monthly)
			for _, m := range months {
				cmd.Printf("  %s: %d\n", m, s.Monthly[m])
			}
		}
		if len(s.Daily) > 0 {
			cmd.Println("\nDaily (last 30):")
			days := sortedKeys(s.Daily)
			for _, d := range days {
				bar := strings.Repeat("|", s.Daily[d])
				cmd.Printf("  %s: %3d %s\n", d, s.Daily[d], bar)
			}
		}
		return nil
	},
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

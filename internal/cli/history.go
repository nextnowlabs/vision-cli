package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

var (
	historyLimit  int
	historySearch string
)

var historyCmd = &cobra.Command{
	Use:   "history [record_id]",
	Short: "View generation history",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			rec, err := appStore.GetRecord(args[0])
			if err != nil {
				cmd.PrintErrf("Record not found: %s\n", args[0])
				return err
			}
			data, _ := json.MarshalIndent(rec, "", "  ")
			cmd.Println(string(data))
			return nil
		}
		records, err := appStore.GetRecords(historyLimit, historySearch)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			cmd.Println("No records")
			return nil
		}
		for _, r := range records {
			ts := ""
			if len(r.Timestamp) >= 16 {
				ts = r.Timestamp[:16]
			}
			ok := "ok"
			if r.Status != "success" {
				ok = "err"
			}
			mode := r.Mode
			if mode == "" {
				mode = "?"
			}
			imgs := len(r.OutputImages)
			promptShort := r.Prompt
			if len(promptShort) > 50 {
				promptShort = promptShort[:50]
			}
			cmd.Printf("  %s  %s  [%s:%s]  %dimg  %s\n", r.ID, ts, mode, ok, imgs, promptShort)
		}
		return nil
	},
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "Number of records to show")
	historyCmd.Flags().StringVarP(&historySearch, "search", "s", "", "Filter by prompt keyword")
}

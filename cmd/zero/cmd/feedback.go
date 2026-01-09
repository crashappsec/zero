package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/crashappsec/zero/pkg/core/config"
	"github.com/crashappsec/zero/pkg/core/feedback"
	"github.com/crashappsec/zero/pkg/core/findings"
	"github.com/spf13/cobra"
)

var feedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Manage analyst feedback on findings",
	Long: `Manage feedback on security findings for rule improvement.

Use this command to mark findings as false positives or true positives,
which helps improve detection accuracy over time.

Examples:
  zero feedback add --fingerprint abc123 --verdict false_positive --reason "Test file"
  zero feedback list --verdict false_positive
  zero feedback stats
  zero feedback export --format csv`,
}

var feedbackAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add feedback for a finding",
	Long: `Add analyst feedback for a specific finding.

The fingerprint uniquely identifies the finding. You can find it in the
scan results JSON output.

Examples:
  zero feedback add --fingerprint abc123 --verdict false_positive --reason "Example code"
  zero feedback add --fingerprint def456 --verdict true_positive`,
	RunE: runFeedbackAdd,
}

var feedbackListCmd = &cobra.Command{
	Use:   "list",
	Short: "List feedback entries",
	Long: `List all feedback entries with optional filters.

Examples:
  zero feedback list                         List all feedback
  zero feedback list --verdict false_positive Filter by verdict
  zero feedback list --rule-id semgrep.secrets.aws-access-key
  zero feedback list --repo expressjs/express`,
	RunE: runFeedbackList,
}

var feedbackStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show feedback statistics",
	Long: `Display aggregate statistics about feedback.

Shows total counts, false positive rates, and rules with high FP rates.`,
	RunE: runFeedbackStats,
}

var feedbackExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export feedback data",
	Long: `Export feedback data for analysis or rule training.

Examples:
  zero feedback export --format csv   Export as CSV
  zero feedback export --format json  Export as JSON`,
	RunE: runFeedbackExport,
}

// Command flags
var (
	feedbackFingerprint string
	feedbackVerdict     string
	feedbackReason      string
	feedbackCategory    string
	feedbackRuleID      string
	feedbackRepo        string
	feedbackFormat      string
	feedbackJSON        bool
	feedbackFPThreshold float64
)

func init() {
	rootCmd.AddCommand(feedbackCmd)

	// Add subcommands
	feedbackCmd.AddCommand(feedbackAddCmd)
	feedbackCmd.AddCommand(feedbackListCmd)
	feedbackCmd.AddCommand(feedbackStatsCmd)
	feedbackCmd.AddCommand(feedbackExportCmd)

	// Add flags
	feedbackAddCmd.Flags().StringVar(&feedbackFingerprint, "fingerprint", "", "Finding fingerprint (required)")
	feedbackAddCmd.Flags().StringVar(&feedbackVerdict, "verdict", "", "Verdict: false_positive, true_positive, needs_review, ignored (required)")
	feedbackAddCmd.Flags().StringVar(&feedbackReason, "reason", "", "Reason for the verdict")
	feedbackAddCmd.Flags().StringVar(&feedbackCategory, "category", "", "Category: test_code, example, documentation, etc.")
	_ = feedbackAddCmd.MarkFlagRequired("fingerprint")
	_ = feedbackAddCmd.MarkFlagRequired("verdict")

	feedbackListCmd.Flags().StringVar(&feedbackVerdict, "verdict", "", "Filter by verdict")
	feedbackListCmd.Flags().StringVar(&feedbackRuleID, "rule-id", "", "Filter by rule ID")
	feedbackListCmd.Flags().StringVar(&feedbackRepo, "repo", "", "Filter by repo (owner/repo)")
	feedbackListCmd.Flags().BoolVar(&feedbackJSON, "json", false, "Output as JSON")

	feedbackStatsCmd.Flags().Float64Var(&feedbackFPThreshold, "fp-threshold", 0.3, "False positive rate threshold for flagging rules")
	feedbackStatsCmd.Flags().BoolVar(&feedbackJSON, "json", false, "Output as JSON")

	feedbackExportCmd.Flags().StringVar(&feedbackFormat, "format", "json", "Export format: csv or json")
}

func runFeedbackAdd(cmd *cobra.Command, args []string) error {
	// Validate verdict
	var verdict feedback.Verdict
	switch strings.ToLower(feedbackVerdict) {
	case "false_positive", "fp":
		verdict = feedback.VerdictFalsePositive
	case "true_positive", "tp":
		verdict = feedback.VerdictTruePositive
	case "needs_review", "review":
		verdict = feedback.VerdictNeedsReview
	case "ignored", "ignore":
		verdict = feedback.VerdictIgnored
	default:
		return fmt.Errorf("invalid verdict: %s (use: false_positive, true_positive, needs_review, ignored)", feedbackVerdict)
	}

	// Load config to get zeroHome
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	// Create storage
	storage := feedback.NewStorage(zeroHome)

	// Create evidence with fingerprint
	evidence := &findings.Evidence{
		Fingerprint: feedbackFingerprint,
	}

	// Create feedback
	fb := feedback.NewFeedback(evidence, verdict, feedbackReason)
	if feedbackCategory != "" {
		fb.Category = feedbackCategory
	}

	// Save
	if err := storage.AddFeedback(fb); err != nil {
		return fmt.Errorf("saving feedback: %w", err)
	}

	fmt.Printf("Feedback added for fingerprint %s\n", feedbackFingerprint)
	fmt.Printf("  Verdict: %s\n", verdict)
	if feedbackReason != "" {
		fmt.Printf("  Reason: %s\n", feedbackReason)
	}

	return nil
}

func runFeedbackList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	storage := feedback.NewStorage(zeroHome)

	// Build query
	query := feedback.FeedbackQuery{}
	if feedbackVerdict != "" {
		query.Verdict = feedback.Verdict(feedbackVerdict)
	}
	if feedbackRuleID != "" {
		query.RuleID = feedbackRuleID
	}
	if feedbackRepo != "" {
		parts := strings.Split(feedbackRepo, "/")
		if len(parts) == 2 {
			query.GitHubOrg = parts[0]
			query.GitHubRepo = parts[1]
		}
	}

	results, err := storage.QueryFeedback(query)
	if err != nil {
		return fmt.Errorf("querying feedback: %w", err)
	}

	if feedbackJSON {
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(results) == 0 {
		fmt.Println("No feedback entries found")
		return nil
	}

	fmt.Printf("Found %d feedback entries:\n\n", len(results))
	for _, fb := range results {
		fmt.Printf("Fingerprint: %s\n", fb.Fingerprint)
		fmt.Printf("  Verdict: %s\n", fb.Verdict)
		if fb.Reason != "" {
			fmt.Printf("  Reason: %s\n", fb.Reason)
		}
		if fb.Evidence != nil && fb.Evidence.RuleID != "" {
			fmt.Printf("  Rule: %s\n", fb.Evidence.RuleID)
		}
		if fb.Evidence != nil && fb.Evidence.FilePath != "" {
			fmt.Printf("  File: %s:%d\n", fb.Evidence.FilePath, fb.Evidence.LineStart)
		}
		fmt.Printf("  Created: %s\n", fb.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

func runFeedbackStats(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	storage := feedback.NewStorage(zeroHome)

	stats, err := storage.GetStats()
	if err != nil {
		return fmt.Errorf("getting stats: %w", err)
	}

	if feedbackJSON {
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println("Feedback Statistics")
	fmt.Println("===================")
	fmt.Printf("Total Feedback:      %d\n", stats.TotalFeedback)
	fmt.Printf("False Positives:     %d\n", stats.FalsePositives)
	fmt.Printf("True Positives:      %d\n", stats.TruePositives)
	fmt.Printf("Needs Review:        %d\n", stats.NeedsReview)
	fmt.Printf("Ignored:             %d\n", stats.Ignored)
	fmt.Printf("False Positive Rate: %.1f%%\n", stats.FalsePositiveRate*100)

	// Show rules with high FP rates
	fpRules, err := storage.GetFalsePositiveRules(feedbackFPThreshold)
	if err != nil {
		return fmt.Errorf("getting FP rules: %w", err)
	}

	if len(fpRules) > 0 {
		fmt.Printf("\nRules with FP rate >= %.0f%%:\n", feedbackFPThreshold*100)
		for _, r := range fpRules {
			fmt.Printf("  %s: %.1f%% FP rate (%d/%d)\n", r.RuleID, r.FPRate*100, r.FalsePositives, r.Total)
		}
	}

	return nil
}

func runFeedbackExport(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	zeroHome := cfg.ZeroHome()
	if zeroHome == "" {
		zeroHome = ".zero"
	}

	storage := feedback.NewStorage(zeroHome)

	var path string
	switch strings.ToLower(feedbackFormat) {
	case "csv":
		path, err = storage.ExportCSV()
	case "json":
		path, err = storage.ExportJSON()
	default:
		return fmt.Errorf("unsupported format: %s (use: csv or json)", feedbackFormat)
	}

	if err != nil {
		return fmt.Errorf("exporting feedback: %w", err)
	}

	// Print the file content
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading export: %w", err)
	}

	fmt.Println(string(data))
	fmt.Fprintf(os.Stderr, "\nExported to: %s\n", path)

	return nil
}

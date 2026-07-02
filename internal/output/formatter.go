package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/driftctl/driftctl/internal/model"
)

// Format outputs a drift report in the specified format.
func Format(w io.Writer, report *model.DriftReport, format string) error {
	switch strings.ToLower(format) {
	case "json":
		return formatJSON(w, report)
	case "table":
		return formatTable(w, report)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// HasDrift returns true if the report contains any findings.
func HasDrift(report *model.DriftReport) bool {
	return report != nil && report.Summary.TotalFindings > 0
}

func formatJSON(w io.Writer, report *model.DriftReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func formatTable(w io.Writer, report *model.DriftReport) error {
	fmt.Fprintf(w, "╔════════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(w, "║ Drift Report: %s\n", report.ScanID)
	fmt.Fprintf(w, "║ Workspace: %s | Status: %s\n", report.Workspace, report.Status)
	fmt.Fprintf(w, "║ Started: %s | Completed: %s\n", report.StartedAt.Format("2006-01-02 15:04:05"), report.CompletedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "╚════════════════════════════════════════════════════════════════╝\n\n")

	// Summary section
	s := report.Summary
	fmt.Fprintf(w, "📊 Summary:\n")
	fmt.Fprintf(w, "  Total Resources: %d\n", s.TotalResources)
	fmt.Fprintf(w, "  Total Findings: %d\n", s.TotalFindings)
	fmt.Fprintf(w, "    - Missing in Cloud: %d\n", s.MissingInCloud)
	fmt.Fprintf(w, "    - Extra in Cloud: %d\n", s.ExtraInCloud)
	fmt.Fprintf(w, "    - Attribute Changes: %d\n", s.AttributeChanges)
	fmt.Fprintf(w, "    - Tag Changes: %d\n\n", s.TagChanges)

	// Findings section
	if len(report.Findings) > 0 {
		fmt.Fprintf(w, "🔍 Findings:\n")
		fmt.Fprintf(w, "┌────────────────────┬──────────────────────┬────────────────────────────────────────┐\n")
		fmt.Fprintf(w, "│ Kind               │ Severity             │ Resource                               │\n")
		fmt.Fprintf(w, "├────────────────────┼──────────────────────┼────────────────────────────────────────┤\n")

		for _, f := range report.Findings {
			resDesc := f.ResourceName
			if resDesc == "" {
				resDesc = f.ResourceID
			}
			if len(resDesc) > 36 {
				resDesc = resDesc[:33] + "..."
			}
			fmt.Fprintf(w, "│ %-18s │ %-20s │ %-36s │\n", f.Kind, f.Severity, resDesc)
			if f.Field != "" {
				fmt.Fprintf(w, "│                    │ Field: %s\n", f.Field)
			}
			if f.Expected != nil || f.Actual != nil {
				expStr := fmt.Sprintf("%v", f.Expected)
				actStr := fmt.Sprintf("%v", f.Actual)
				if len(expStr) > 40 {
					expStr = expStr[:37] + "..."
				}
				if len(actStr) > 40 {
					actStr = actStr[:37] + "..."
				}
				fmt.Fprintf(w, "│                    │ Expected: %s\n", expStr)
				fmt.Fprintf(w, "│                    │ Actual:   %s\n", actStr)
			}
		}
		fmt.Fprintf(w, "└────────────────────┴──────────────────────┴────────────────────────────────────────┘\n")
	} else {
		fmt.Fprintf(w, "✅ No drift detected!\n")
	}

	// Errors section
	if len(report.Errors) > 0 {
		fmt.Fprintf(w, "\n⚠️  Errors:\n")
		for i, err := range report.Errors {
			fmt.Fprintf(w, "  %d. %s\n", i+1, err)
		}
	}

	return nil
}

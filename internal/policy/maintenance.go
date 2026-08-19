package policy

import (
	"fmt"
	"sort"
	"strings"
)

type MaintenanceRecord struct {
	Yard      string
	ConsistID string
	DueHours  float64
	Active    bool
	Tags      []string
	Revision  int
}

type MaintenanceDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateMaintenance(r MaintenanceRecord) MaintenanceDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if r.DueHours <= 0 {
		reasons = append(reasons, "DueHours is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeMaintenanceTags(r.Tags)
	score := 100 - len(reasons)*20 + minMaintenance(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return MaintenanceDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateMaintenanceBatch(records []MaintenanceRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.DueHours)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate maintenance record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateMaintenance(r); !d.Allowed {
			return fmt.Errorf("invalid maintenance record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeMaintenanceTags(tags []string) []string {
	unique := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" {
			unique[tag] = struct{}{}
		}
	}
	out := make([]string, 0, len(unique))
	for tag := range unique {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}
func minMaintenance(a, b int) int {
	if a < b {
		return a
	}
	return b
}

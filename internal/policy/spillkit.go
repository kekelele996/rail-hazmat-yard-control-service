package policy

import (
	"fmt"
	"sort"
	"strings"
)

type SpillKitRecord struct {
	Yard      string
	ConsistID string
	Units     float64
	Active    bool
	Tags      []string
	Revision  int
}

type SpillKitDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateSpillKit(r SpillKitRecord) SpillKitDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if r.Units <= 0 {
		reasons = append(reasons, "Units is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeSpillKitTags(r.Tags)
	score := 100 - len(reasons)*20 + minSpillKit(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return SpillKitDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateSpillKitBatch(records []SpillKitRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.Units)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate spillkit record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateSpillKit(r); !d.Allowed {
			return fmt.Errorf("invalid spillkit record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeSpillKitTags(tags []string) []string {
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
func minSpillKit(a, b int) int {
	if a < b {
		return a
	}
	return b
}

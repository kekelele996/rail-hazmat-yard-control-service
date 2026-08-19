package policy

import (
	"fmt"
	"sort"
	"strings"
)

type GradientRecord struct {
	Yard      string
	ConsistID string
	Permille  float64
	Active    bool
	Tags      []string
	Revision  int
}

type GradientDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateGradient(r GradientRecord) GradientDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if r.Permille <= 0 {
		reasons = append(reasons, "Permille is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeGradientTags(r.Tags)
	score := 100 - len(reasons)*20 + minGradient(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return GradientDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateGradientBatch(records []GradientRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.Permille)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate gradient record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateGradient(r); !d.Allowed {
			return fmt.Errorf("invalid gradient record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeGradientTags(tags []string) []string {
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
func minGradient(a, b int) int {
	if a < b {
		return a
	}
	return b
}

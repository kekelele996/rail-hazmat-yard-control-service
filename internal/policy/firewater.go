package policy

import (
	"fmt"
	"sort"
	"strings"
)

type FirewaterRecord struct {
	Yard      string
	ConsistID string
	Pressure  float64
	Active    bool
	Tags      []string
	Revision  int
}

type FirewaterDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateFirewater(r FirewaterRecord) FirewaterDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if r.Pressure <= 0 {
		reasons = append(reasons, "Pressure is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeFirewaterTags(r.Tags)
	score := 100 - len(reasons)*20 + minFirewater(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return FirewaterDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateFirewaterBatch(records []FirewaterRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.Pressure)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate firewater record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateFirewater(r); !d.Allowed {
			return fmt.Errorf("invalid firewater record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeFirewaterTags(tags []string) []string {
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
func minFirewater(a, b int) int {
	if a < b {
		return a
	}
	return b
}

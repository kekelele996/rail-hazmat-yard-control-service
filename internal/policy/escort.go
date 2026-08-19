package policy

import (
	"fmt"
	"sort"
	"strings"
)

type EscortRecord struct {
	Yard      string
	ConsistID string
	EscortID  string
	Active    bool
	Tags      []string
	Revision  int
}

type EscortDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateEscort(r EscortRecord) EscortDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if strings.TrimSpace(r.EscortID) == "" {
		reasons = append(reasons, "EscortID is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeEscortTags(r.Tags)
	score := 100 - len(reasons)*20 + minEscort(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return EscortDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateEscortBatch(records []EscortRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.EscortID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate escort record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateEscort(r); !d.Allowed {
			return fmt.Errorf("invalid escort record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeEscortTags(tags []string) []string {
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
func minEscort(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package policy

import (
	"fmt"
	"sort"
	"strings"
)

type CrossingRecord struct {
	Yard       string
	ConsistID  string
	CrossingID string
	Active     bool
	Tags       []string
	Revision   int
}

type CrossingDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateCrossing(r CrossingRecord) CrossingDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if strings.TrimSpace(r.CrossingID) == "" {
		reasons = append(reasons, "CrossingID is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeCrossingTags(r.Tags)
	score := 100 - len(reasons)*20 + minCrossing(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return CrossingDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateCrossingBatch(records []CrossingRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.CrossingID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate crossing record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateCrossing(r); !d.Allowed {
			return fmt.Errorf("invalid crossing record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeCrossingTags(tags []string) []string {
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
func minCrossing(a, b int) int {
	if a < b {
		return a
	}
	return b
}

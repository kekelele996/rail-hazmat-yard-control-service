package policy

import (
	"fmt"
	"sort"
	"strings"
)

type BrakeRecord struct {
	Yard       string
	ConsistID  string
	BrakeRatio float64
	Active     bool
	Tags       []string
	Revision   int
}

type BrakeDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateBrake(r BrakeRecord) BrakeDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if r.BrakeRatio <= 0 {
		reasons = append(reasons, "BrakeRatio is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeBrakeTags(r.Tags)
	score := 100 - len(reasons)*20 + minBrake(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return BrakeDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateBrakeBatch(records []BrakeRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.BrakeRatio)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate brake record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateBrake(r); !d.Allowed {
			return fmt.Errorf("invalid brake record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeBrakeTags(tags []string) []string {
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
func minBrake(a, b int) int {
	if a < b {
		return a
	}
	return b
}

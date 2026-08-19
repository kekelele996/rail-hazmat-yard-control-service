package policy

import (
	"fmt"
	"sort"
	"strings"
)

type LightingRecord struct {
	Yard      string
	ConsistID string
	Lux       float64
	Active    bool
	Tags      []string
	Revision  int
}

type LightingDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateLighting(r LightingRecord) LightingDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if r.Lux <= 0 {
		reasons = append(reasons, "Lux is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeLightingTags(r.Tags)
	score := 100 - len(reasons)*20 + minLighting(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return LightingDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateLightingBatch(records []LightingRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.Lux)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate lighting record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateLighting(r); !d.Allowed {
			return fmt.Errorf("invalid lighting record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeLightingTags(tags []string) []string {
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
func minLighting(a, b int) int {
	if a < b {
		return a
	}
	return b
}

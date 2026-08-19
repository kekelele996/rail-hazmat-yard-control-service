package policy

import (
	"fmt"
	"sort"
	"strings"
)

type PlacardRecord struct {
	Yard        string
	ConsistID   string
	PlacardCode string
	Active      bool
	Tags        []string
	Revision    int
}

type PlacardDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluatePlacard(r PlacardRecord) PlacardDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if strings.TrimSpace(r.PlacardCode) == "" {
		reasons = append(reasons, "PlacardCode is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizePlacardTags(r.Tags)
	score := 100 - len(reasons)*20 + minPlacard(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return PlacardDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidatePlacardBatch(records []PlacardRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.PlacardCode)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate placard record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluatePlacard(r); !d.Allowed {
			return fmt.Errorf("invalid placard record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizePlacardTags(tags []string) []string {
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
func minPlacard(a, b int) int {
	if a < b {
		return a
	}
	return b
}

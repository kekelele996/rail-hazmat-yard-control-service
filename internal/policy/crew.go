package policy

import (
	"fmt"
	"sort"
	"strings"
)

type CrewRecord struct {
	Yard          string
	ConsistID     string
	Certification string
	Active        bool
	Tags          []string
	Revision      int
}

type CrewDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateCrew(r CrewRecord) CrewDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if strings.TrimSpace(r.Certification) == "" {
		reasons = append(reasons, "Certification is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeCrewTags(r.Tags)
	score := 100 - len(reasons)*20 + minCrew(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return CrewDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateCrewBatch(records []CrewRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.Certification)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate crew record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateCrew(r); !d.Allowed {
			return fmt.Errorf("invalid crew record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeCrewTags(tags []string) []string {
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
func minCrew(a, b int) int {
	if a < b {
		return a
	}
	return b
}

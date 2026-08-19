package policy

import (
	"fmt"
	"sort"
	"strings"
)

type InsuranceRecord struct {
	Yard      string
	ConsistID string
	PolicyID  string
	Active    bool
	Tags      []string
	Revision  int
}

type InsuranceDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateInsurance(r InsuranceRecord) InsuranceDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if strings.TrimSpace(r.PolicyID) == "" {
		reasons = append(reasons, "PolicyID is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeInsuranceTags(r.Tags)
	score := 100 - len(reasons)*20 + minInsurance(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return InsuranceDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateInsuranceBatch(records []InsuranceRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.PolicyID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate insurance record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateInsurance(r); !d.Allowed {
			return fmt.Errorf("invalid insurance record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeInsuranceTags(tags []string) []string {
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
func minInsurance(a, b int) int {
	if a < b {
		return a
	}
	return b
}

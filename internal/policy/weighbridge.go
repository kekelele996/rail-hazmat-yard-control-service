package policy

import (
	"fmt"
	"sort"
	"strings"
)

type WeighbridgeRecord struct {
	Yard      string
	ConsistID string
	TicketID  string
	Active    bool
	Tags      []string
	Revision  int
}

type WeighbridgeDecision struct {
	Allowed        bool
	Reasons        []string
	NormalizedTags []string
	Score          int
}

func EvaluateWeighbridge(r WeighbridgeRecord) WeighbridgeDecision {
	reasons := make([]string, 0, 4)
	if strings.TrimSpace(r.Yard) == "" {
		reasons = append(reasons, "yard is required")
	}
	if strings.TrimSpace(r.ConsistID) == "" {
		reasons = append(reasons, "consist is required")
	}
	if strings.TrimSpace(r.TicketID) == "" {
		reasons = append(reasons, "TicketID is required")
	}
	if !r.Active {
		reasons = append(reasons, "record is inactive")
	}
	tags := normalizeWeighbridgeTags(r.Tags)
	score := 100 - len(reasons)*20 + minWeighbridge(len(tags)*2, 10)
	if score < 0 {
		score = 0
	}
	return WeighbridgeDecision{Allowed: len(reasons) == 0, Reasons: reasons, NormalizedTags: tags, Score: score}
}

func ValidateWeighbridgeBatch(records []WeighbridgeRecord) error {
	seen := make(map[string]struct{}, len(records))
	for i, r := range records {
		key := fmt.Sprintf("%s/%s/%v", strings.TrimSpace(r.Yard), strings.TrimSpace(r.ConsistID), r.TicketID)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate weighbridge record at index %d", i)
		}
		seen[key] = struct{}{}
		if d := EvaluateWeighbridge(r); !d.Allowed {
			return fmt.Errorf("invalid weighbridge record %d: %s", i, strings.Join(d.Reasons, ", "))
		}
	}
	return nil
}

func normalizeWeighbridgeTags(tags []string) []string {
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
func minWeighbridge(a, b int) int {
	if a < b {
		return a
	}
	return b
}

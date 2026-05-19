package tools

import (
	"testing"

	"ariadne/internal/model"
)

func TestBuildConstraintsFromRequestDistinguishesHomestayIntent(t *testing.T) {
	tool := NewPreferenceConstraintTool()

	preferConstraints := tool.BuildConstraintsFromRequest(model.TripRequest{
		RawInput: "想体验海边民宿，最好有特色",
	})
	if !hasHotelPreferTag(preferConstraints, "homestay") {
		t.Fatalf("expected homestay preference, got %#v", preferConstraints)
	}
	if hasHotelHardAvoidTag(preferConstraints, "homestay") {
		t.Fatalf("did not expect hard avoid homestay for positive request: %#v", preferConstraints)
	}

	avoidConstraints := tool.BuildConstraintsFromRequest(model.TripRequest{
		RawInput: "不要民宿，不住客栈",
	})
	if !hasHotelHardAvoidTag(avoidConstraints, "homestay") || !hasHotelHardAvoidTag(avoidConstraints, "guesthouse") {
		t.Fatalf("expected hard avoid homestay and guesthouse, got %#v", avoidConstraints)
	}
}

func hasHotelPreferTag(constraints []model.PreferenceConstraint, tag string) bool {
	for _, constraint := range constraints {
		if constraint.Domain != "hotel" {
			continue
		}
		if hasTag(constraint.PreferTags, tag) {
			return true
		}
	}

	return false
}

func hasHotelHardAvoidTag(constraints []model.PreferenceConstraint, tag string) bool {
	for _, constraint := range constraints {
		if constraint.Domain != "hotel" || constraint.Strength != "hard" {
			continue
		}
		if hasTag(constraint.AvoidTags, tag) {
			return true
		}
	}

	return false
}

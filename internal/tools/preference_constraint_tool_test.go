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

func TestResolvePreferenceConstraintsCurrentHomestayOverridesLongTermAvoid(t *testing.T) {
	profile := ResolvePreferenceConstraints([]model.PreferenceConstraint{
		{
			Domain:    "hotel",
			AvoidTags: []string{"homestay", "guesthouse", "hostel"},
			Strength:  "hard",
			Priority:  60,
			Source:    "long_term_memory",
		},
		{
			Domain:     "hotel",
			PreferTags: []string{"homestay", "guesthouse", "sea_nearby", "unique_stay"},
			Strength:   "soft",
			Priority:   95,
			Source:     "current_request",
		},
	})

	if hasString(profile.Hotel.HardAvoidTags, "homestay") || hasString(profile.Hotel.HardAvoidTags, "guesthouse") {
		t.Fatalf("current homestay/guesthouse preference should override long-term avoid: %#v", profile.Hotel)
	}

	if !hasString(profile.Hotel.HardAvoidTags, "hostel") {
		t.Fatalf("hostel hard avoid should remain: %#v", profile.Hotel)
	}
}

func TestBuildConstraintsFromRequestTreatsShoppingNegationAsHardAvoid(t *testing.T) {
	tool := NewPreferenceConstraintTool()

	constraints := tool.BuildConstraintsFromRequest(model.TripRequest{
		RawInput: "不想购物，不要普通商场，也不要市场",
	})

	if !hasAttractionHardAvoidTag(constraints, "shopping") ||
		!hasAttractionHardAvoidTag(constraints, "commercial_area") {
		t.Fatalf("expected hard avoid shopping/commercial_area, got %#v", constraints)
	}

	for _, constraint := range constraints {
		if constraint.Domain == "attraction" && hasTag(constraint.PreferTags, "shopping") {
			t.Fatalf("shopping negation must not become prefer shopping: %#v", constraints)
		}
	}
}

func hasAttractionHardAvoidTag(constraints []model.PreferenceConstraint, tag string) bool {
	for _, constraint := range constraints {
		if constraint.Domain != "attraction" || constraint.Strength != "hard" {
			continue
		}
		if hasTag(constraint.AvoidTags, tag) {
			return true
		}
	}

	return false
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

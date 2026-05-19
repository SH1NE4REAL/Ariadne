package tools

import (
	"strings"

	"ariadne/internal/model"
)

func BuildTripIntentProfile(constraints []model.PreferenceConstraint, source string) model.TripIntentProfile {
	profile := model.TripIntentProfile{}

	for _, constraint := range constraints {
		if source != "" && strings.ToLower(strings.TrimSpace(constraint.Source)) != source {
			continue
		}

		addConstraintToTripIntentProfile(&profile, constraint)
	}

	return profile
}

func TripIntentProfileHasPOISearchIntent(profile model.TripIntentProfile) bool {
	return domainIntentHasSignal(profile.AttractionIntent) ||
		domainIntentHasSignal(profile.FoodIntent) ||
		domainIntentHasSignal(profile.RouteIntent)
}

func addConstraintToTripIntentProfile(profile *model.TripIntentProfile, constraint model.PreferenceConstraint) {
	domains := constraintDomainsForResolve(constraint.Domain)

	for _, domain := range domains {
		intent := tripIntentDomain(profile, domain)
		addConstraintToDomainIntent(&intent, constraint)
		setTripIntentDomain(profile, domain, intent)
	}
}

func addConstraintToDomainIntent(intent *model.DomainIntent, constraint model.PreferenceConstraint) {
	hard := strings.ToLower(strings.TrimSpace(constraint.Strength)) == "hard"
	source := strings.ToLower(strings.TrimSpace(constraint.Source))
	if intent.Source == "" {
		intent.Source = source
	}

	if hard {
		for _, tag := range constraint.PreferTags {
			addUniqueLower(&intent.HardPreferTags, tag)
		}
		for _, keyword := range constraint.PreferKeywords {
			addUniqueLower(&intent.HardPreferKeywords, keyword)
		}
		for _, tag := range constraint.AvoidTags {
			addUniqueLower(&intent.HardAvoidTags, tag)
		}
		for _, keyword := range constraint.ExcludeKeywords {
			addUniqueLower(&intent.HardAvoidKeywords, keyword)
		}
		return
	}

	for _, tag := range constraint.PreferTags {
		addUniqueLower(&intent.SoftPreferTags, tag)
	}
	for _, keyword := range constraint.PreferKeywords {
		addUniqueLower(&intent.SoftPreferKeywords, keyword)
	}
	for _, tag := range constraint.AvoidTags {
		addUniqueLower(&intent.SoftAvoidTags, tag)
	}
	for _, keyword := range constraint.ExcludeKeywords {
		addUniqueLower(&intent.SoftAvoidKeywords, keyword)
	}
}

func tripIntentDomain(profile *model.TripIntentProfile, domain string) model.DomainIntent {
	switch domain {
	case "transport":
		return profile.TransportIntent
	case "hotel":
		return profile.HotelIntent
	case "attraction":
		return profile.AttractionIntent
	case "food":
		return profile.FoodIntent
	case "route":
		return profile.RouteIntent
	default:
		return profile.ConstraintIntent
	}
}

func setTripIntentDomain(profile *model.TripIntentProfile, domain string, intent model.DomainIntent) {
	switch domain {
	case "transport":
		profile.TransportIntent = intent
	case "hotel":
		profile.HotelIntent = intent
	case "attraction":
		profile.AttractionIntent = intent
	case "food":
		profile.FoodIntent = intent
	case "route":
		profile.RouteIntent = intent
	default:
		profile.ConstraintIntent = intent
	}
}

func domainIntentHasSignal(intent model.DomainIntent) bool {
	return len(intent.HardPreferTags) > 0 ||
		len(intent.SoftPreferTags) > 0 ||
		len(intent.HardAvoidTags) > 0 ||
		len(intent.SoftAvoidTags) > 0 ||
		len(intent.HardPreferKeywords) > 0 ||
		len(intent.SoftPreferKeywords) > 0 ||
		len(intent.HardAvoidKeywords) > 0 ||
		len(intent.SoftAvoidKeywords) > 0
}

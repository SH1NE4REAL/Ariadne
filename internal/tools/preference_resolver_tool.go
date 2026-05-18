package tools

import (
	"sort"
	"strings"

	"ariadne/internal/model"
)

func ResolvePreferenceConstraints(constraints []model.PreferenceConstraint) model.EffectivePreferenceProfile {
	sorted := make([]model.PreferenceConstraint, len(constraints))
	copy(sorted, constraints)

	sort.SliceStable(sorted, func(i, j int) bool {
		return constraintResolvePriority(sorted[i]) > constraintResolvePriority(sorted[j])
	})

	profile := model.EffectivePreferenceProfile{}

	for _, constraint := range sorted {
		addConstraintToProfile(&profile, constraint)
	}

	return profile
}

func addConstraintToProfile(profile *model.EffectivePreferenceProfile, constraint model.PreferenceConstraint) {
	domains := constraintDomainsForResolve(constraint.Domain)

	for _, domain := range domains {
		preference := profile.DomainPreference(domain)
		addConstraintToDomainPreference(&preference, constraint)
		setDomainPreference(profile, domain, preference)
	}
}

func addConstraintToDomainPreference(preference *model.EffectiveDomainPreference, constraint model.PreferenceConstraint) {
	hard := strings.ToLower(strings.TrimSpace(constraint.Strength)) == "hard"

	if hard {
		for _, tag := range constraint.AvoidTags {
			addUniqueLower(&preference.HardAvoidTags, tag)
			removeString(&preference.HardPreferTags, tag)
			removeString(&preference.SoftPreferTags, tag)
		}

		for _, keyword := range constraint.ExcludeKeywords {
			addUniqueLower(&preference.HardAvoidKeywords, keyword)
			removeString(&preference.HardPreferKeywords, keyword)
			removeString(&preference.SoftPreferKeywords, keyword)
		}

		for _, tag := range constraint.PreferTags {
			if hasString(preference.HardAvoidTags, tag) {
				continue
			}
			addUniqueLower(&preference.HardPreferTags, tag)
		}

		for _, keyword := range constraint.PreferKeywords {
			if hasString(preference.HardAvoidKeywords, keyword) {
				continue
			}
			addUniqueLower(&preference.HardPreferKeywords, keyword)
		}

		return
	}

	for _, tag := range constraint.AvoidTags {
		if hasString(preference.HardPreferTags, tag) {
			continue
		}
		addUniqueLower(&preference.SoftAvoidTags, tag)
	}

	for _, keyword := range constraint.ExcludeKeywords {
		if hasString(preference.HardPreferKeywords, keyword) {
			continue
		}
		addUniqueLower(&preference.SoftAvoidKeywords, keyword)
	}

	for _, tag := range constraint.PreferTags {
		if hasString(preference.HardAvoidTags, tag) || hasString(preference.SoftAvoidTags, tag) {
			continue
		}
		addUniqueLower(&preference.SoftPreferTags, tag)
	}

	for _, keyword := range constraint.PreferKeywords {
		if hasString(preference.HardAvoidKeywords, keyword) || hasString(preference.SoftAvoidKeywords, keyword) {
			continue
		}
		addUniqueLower(&preference.SoftPreferKeywords, keyword)
	}
}

func constraintResolvePriority(constraint model.PreferenceConstraint) int {
	priority := constraint.Priority
	if priority <= 0 {
		if strings.ToLower(strings.TrimSpace(constraint.Source)) == "current_request" {
			priority = 80
		} else {
			priority = 40
		}
	}

	if strings.ToLower(strings.TrimSpace(constraint.Source)) == "current_request" {
		priority += 1000
	}

	if strings.ToLower(strings.TrimSpace(constraint.Strength)) == "hard" {
		priority += 100
	}

	return priority
}

func constraintDomainsForResolve(domain string) []string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" || domain == "general" {
		return []string{"transport", "hotel", "attraction", "route", "food"}
	}

	return []string{domain}
}

func setDomainPreference(profile *model.EffectivePreferenceProfile, domain string, preference model.EffectiveDomainPreference) {
	switch domain {
	case "transport":
		profile.Transport = preference
	case "hotel":
		profile.Hotel = preference
	case "attraction":
		profile.Attraction = preference
	case "route":
		profile.Route = preference
	case "food":
		profile.Food = preference
	}
}

func addUniqueLower(items *[]string, value string) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || hasString(*items, value) {
		return
	}

	*items = append(*items, value)
}

func removeString(items *[]string, value string) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return
	}

	result := (*items)[:0]
	for _, item := range *items {
		if strings.ToLower(strings.TrimSpace(item)) == value {
			continue
		}
		result = append(result, item)
	}

	*items = result
}

func hasString(items []string, value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item)) == value {
			return true
		}
	}

	return false
}

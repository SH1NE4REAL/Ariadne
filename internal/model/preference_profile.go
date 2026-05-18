package model

type EffectivePreferenceProfile struct {
	Transport  EffectiveDomainPreference `json:"transport"`
	Hotel      EffectiveDomainPreference `json:"hotel"`
	Attraction EffectiveDomainPreference `json:"attraction"`
	Route      EffectiveDomainPreference `json:"route"`
	Food       EffectiveDomainPreference `json:"food"`
}

type EffectiveDomainPreference struct {
	HardPreferTags []string `json:"hard_prefer_tags"`
	SoftPreferTags []string `json:"soft_prefer_tags"`
	HardAvoidTags  []string `json:"hard_avoid_tags"`
	SoftAvoidTags  []string `json:"soft_avoid_tags"`

	HardPreferKeywords []string `json:"hard_prefer_keywords"`
	SoftPreferKeywords []string `json:"soft_prefer_keywords"`
	HardAvoidKeywords  []string `json:"hard_avoid_keywords"`
	SoftAvoidKeywords  []string `json:"soft_avoid_keywords"`
}

func (p EffectivePreferenceProfile) DomainPreference(domain string) EffectiveDomainPreference {
	switch domain {
	case "transport":
		return p.Transport
	case "hotel":
		return p.Hotel
	case "attraction":
		return p.Attraction
	case "route":
		return p.Route
	case "food":
		return p.Food
	default:
		return EffectiveDomainPreference{}
	}
}

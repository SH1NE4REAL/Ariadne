package tools

import (
	"encoding/json"
	"strings"

	"ariadne/internal/model"
)

type PreferenceConstraintTool struct {
	Name        string
	Description string
}

func NewPreferenceConstraintTool() PreferenceConstraintTool {
	return PreferenceConstraintTool{
		Name:        "preference_constraint_tool",
		Description: "将向量检索到的长期记忆转换为结构化偏好约束",
	}
}

type memoryConstraintPayload struct {
	Constraints      []model.PreferenceConstraint `json:"constraints"`
	Domain           string                       `json:"domain"`
	ExcludeKeywords  []string                     `json:"exclude_keywords"`
	PreferKeywords   []string                     `json:"prefer_keywords"`
	NegativeKeywords []string                     `json:"negative_keywords"`
	PositiveKeywords []string                     `json:"positive_keywords"`
	AvoidTags        []string                     `json:"avoid_tags"`
	PreferTags       []string                     `json:"prefer_tags"`
	Strength         string                       `json:"strength"`
	Reason           string                       `json:"reason"`
	Scenario         string                       `json:"scenario"`
	Source           string                       `json:"source"`
	Priority         int                          `json:"priority"`
}

func (t PreferenceConstraintTool) BuildConstraints(memories []model.MemoryRecord) []model.PreferenceConstraint {
	constraints := make([]model.PreferenceConstraint, 0)

	for _, memory := range memories {
		fromMetadata := parseConstraintsFromMetadata(memory)
		if len(fromMetadata) > 0 {
			constraints = append(constraints, fromMetadata...)
			continue
		}

		constraints = append(constraints, inferConstraintsFromText(memory)...)
	}

	return deduplicateConstraints(normalizeConstraints(constraints))
}

func (t PreferenceConstraintTool) BuildConstraintsFromRequest(request model.TripRequest) []model.PreferenceConstraint {
	text := strings.ToLower(strings.Join([]string{
		request.RawInput,
		request.Preference,
		request.TransportPreference,
		request.LocalTransportMode,
	}, " "))

	constraints := make([]model.PreferenceConstraint, 0)
	avoidMuseum := containsAnyConstraintText(text, []string{"不想逛博物馆", "不想去博物馆", "不要博物馆", "不逛博物馆"})
	avoidFlight := containsAnyConstraintText(text, []string{"不要飞机", "不坐飞机", "别坐飞机", "不要坐飞机", "不想坐飞机", "不喜欢坐飞机"})
	preferHighSpeedTrain := containsAnyConstraintText(text, []string{"优先高铁", "优先动车", "高铁或动车", "高铁动车", "坐高铁", "坐动车"})
	childFriendly := containsAnyConstraintText(text, []string{"亲子", "带孩子", "带小孩", "小朋友", "儿童", "8岁", "八岁"})
	avoidNightlife := containsAnyConstraintText(text, []string{"不要酒吧", "不要夜生活", "不去酒吧", "不去夜店", "不要夜店", "不想夜生活"})
	avoidInfluencerStreet := containsAnyConstraintText(text, []string{"不要网红街区", "不去网红街区", "不要网红店", "不要网红"})
	avoidShopping := containsAnyConstraintText(text, []string{"不要普通商场", "不去普通商场", "不要商场", "不逛商场"})
	avoidHomestay := hasNegativeAccommodationIntent(text, []string{"民宿", "客栈", "公寓", "青旅", "旅社"})
	preferHomestay := !avoidHomestay && hasPositiveAccommodationIntent(text, []string{"民宿", "客栈"})

	if avoidFlight || preferHighSpeedTrain {
		strength := "soft"
		priority := 90
		if avoidFlight {
			strength = "hard"
			priority = 100
		}

		constraints = append(constraints, model.PreferenceConstraint{
			Domain:          "transport",
			PreferTags:      []string{"high_speed_train", "bullet_train", "train"},
			AvoidTags:       []string{"flight"},
			PreferKeywords:  []string{"高铁", "动车"},
			ExcludeKeywords: []string{"飞机", "航班", "机场"},
			Strength:        strength,
			Priority:        priority,
			Source:          "current_request",
			Reason:          "当前请求偏好高铁或动车，并排除飞机",
			SourceMemoryID:  "current_request",
		})
	}

	if avoidHomestay {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:          "hotel",
			PreferTags:      []string{"hotel", "chain_hotel", "metro_nearby", "clean"},
			AvoidTags:       []string{"hostel", "homestay", "apartment", "family_inn", "guesthouse"},
			PreferKeywords:  []string{"酒店", "连锁", "地铁", "如家", "汉庭", "全季", "速8"},
			ExcludeKeywords: []string{"青旅", "青年旅舍", "民宿", "公寓", "家庭旅社", "客栈", "旅社"},
			Strength:        "hard",
			Priority:        100,
			Source:          "current_request",
			Reason:          "当前请求明确排除非标准住宿形态",
			SourceMemoryID:  "current_request",
		})
	}

	if preferHomestay {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "hotel",
			PreferTags:     []string{"homestay", "guesthouse", "local_stay"},
			PreferKeywords: []string{"民宿", "客栈", "特色住宿", "海边民宿"},
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
			Reason:         "当前请求表达了想体验民宿或客栈类特色住宿",
			SourceMemoryID: "current_request",
		})
	}

	if avoidMuseum {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:          "attraction",
			AvoidTags:       []string{"museum", "exhibition", "art_gallery"},
			ExcludeKeywords: []string{"博物馆", "展览馆", "美术馆"},
			Strength:        "hard",
			Priority:        100,
			Source:          "current_request",
			Reason:          "当前请求明确不想逛博物馆或展览馆",
			SourceMemoryID:  "current_request",
		})
	}

	if !avoidMuseum && containsAnyConstraintText(text, []string{"博物馆", "展览", "美术馆", "人文", "历史"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:     "attraction",
			PreferTags: []string{"museum", "exhibition", "art_gallery", "culture", "indoor", "low_exertion"},
			PreferKeywords: []string{
				"博物馆", "展览馆", "美术馆", "纪念馆", "历史", "人文",
			},
			Strength:       "soft",
			Priority:       100,
			Source:         "current_request",
			Reason:         "当前请求偏好文化、博物馆或展览类景点",
			SourceMemoryID: "current_request",
		})
	}

	if containsAnyConstraintText(text, []string{"不想爬山", "不要爬山", "不喜欢爬山", "不要徒步", "不想徒步"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:    "attraction",
			AvoidTags: []string{"mountain", "hiking", "high_exertion"},
			ExcludeKeywords: []string{
				"爬山", "登山", "徒步", "长城", "山", "峡谷", "凤凰岭",
			},
			Strength:       "hard",
			Priority:       100,
			Source:         "current_request",
			Reason:         "当前请求明确不想爬山或徒步",
			SourceMemoryID: "current_request",
		})
	}

	if containsAnyConstraintText(text, []string{"轻松", "慢节奏", "不想太赶", "不要特种兵"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "route",
			PreferTags:     []string{"relaxed", "slow_pace", "short_transfer"},
			AvoidTags:      []string{"tight_schedule", "long_transfer", "too_many_attractions"},
			Strength:       "soft",
			Priority:       100,
			Source:         "current_request",
			Reason:         "当前请求偏好轻松旅行",
			SourceMemoryID: "current_request",
		})

		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"low_exertion", "indoor", "city_walk"},
			AvoidTags:      []string{"high_exertion", "remote"},
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
			Reason:         "当前请求希望降低体力消耗",
			SourceMemoryID: "current_request",
		})
	}

	if childFriendly || containsAnyConstraintText(text, []string{"科技馆", "科学中心", "海洋馆", "水族馆", "室内亲子"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:     "attraction",
			PreferTags: []string{"family", "indoor", "science_museum", "aquarium", "low_exertion"},
			AvoidTags: []string{
				"mountain", "hiking", "high_exertion", "bar", "nightlife",
				"adult_or_couple", "remote",
			},
			PreferKeywords: []string{"科技馆", "科学中心", "海洋馆", "水族馆", "儿童展览", "科普"},
			ExcludeKeywords: []string{
				"失恋博物馆", "酒吧", "夜店", "夜生活", "成人", "情侣", "爬山", "徒步", "长城",
			},
			Strength:       "hard",
			Priority:       100,
			Source:         "current_request",
			Reason:         "当前请求是亲子室内场景，偏好科技馆或海洋馆，并排除成人向、夜生活和高体力景点",
			SourceMemoryID: "current_request",
		})

		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "route",
			PreferTags:     []string{"relaxed", "child_friendly", "short_transfer"},
			AvoidTags:      []string{"tight_schedule", "long_transfer", "too_many_attractions"},
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
			Reason:         "当前请求是亲子场景，路线应轻松短通勤",
			SourceMemoryID: "current_request",
		})
	}

	if avoidNightlife {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:          "attraction",
			AvoidTags:       []string{"bar", "nightlife", "adult_or_couple"},
			ExcludeKeywords: []string{"酒吧", "夜店", "夜生活", "成人", "情侣"},
			Strength:        "hard",
			Priority:        100,
			Source:          "current_request",
			Reason:          "当前请求明确不要酒吧或夜生活",
			SourceMemoryID:  "current_request",
		})
	}

	if avoidInfluencerStreet {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:          "attraction",
			AvoidTags:       []string{"crowded", "influencer_spot"},
			ExcludeKeywords: []string{"网红", "打卡", "网红街区"},
			Strength:        "hard",
			Priority:        100,
			Source:          "current_request",
			Reason:          "当前请求明确不要网红街区",
			SourceMemoryID:  "current_request",
		})
	}

	if avoidShopping {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:          "attraction",
			AvoidTags:       []string{"shopping", "commercial_area"},
			ExcludeKeywords: []string{"普通商场", "商场", "购物中心"},
			Strength:        "hard",
			Priority:        100,
			Source:          "current_request",
			Reason:          "当前请求明确不要普通商场",
			SourceMemoryID:  "current_request",
		})
	}

	if containsAnyConstraintText(text, []string{"不想晒太阳", "不要晒太阳", "怕晒", "少晒", "室内"}) {
		indoorKeywords := []string{"室内景点", "科技馆", "海洋馆", "展览馆"}
		if childFriendly {
			indoorKeywords = []string{"科技馆", "海洋馆", "水族馆", "儿童展览", "科普"}
		}

		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"indoor", "low_exertion"},
			AvoidTags:      []string{"outdoor", "high_exertion"},
			PreferKeywords: indoorKeywords,
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
			Reason:         "当前请求偏好室内或少晒太阳的安排",
			SourceMemoryID: "current_request",
		})
	}

	if containsAnyConstraintText(text, []string{"小吃", "美食", "地道吃的", "本地吃的"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "food",
			PreferTags:     []string{"food", "local_food", "snack_street"},
			PreferKeywords: []string{"小吃", "美食", "本地美食", "夜市"},
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
			Reason:         "当前请求希望体验本地美食",
			SourceMemoryID: "current_request",
		})
	}

	if containsAnyConstraintText(text, []string{"夜景", "晚上逛", "夜游"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"night_view", "landmark", "city_walk"},
			PreferKeywords: []string{"夜景", "观景", "地标"},
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
			Reason:         "当前请求偏好夜景或夜游",
			SourceMemoryID: "current_request",
		})
	}

	if containsAnyConstraintText(text, []string{"看海", "海边", "海滩", "沙滩", "海滨", "海岸", "滨海", "观海"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"sea", "beach", "waterfront", "coast", "low_exertion"},
			PreferKeywords: []string{"海边", "海滩", "沙滩", "海滨", "海岸", "滨海步道", "观海"},
			Strength:       "soft",
			Priority:       95,
			Source:         "current_request",
			Reason:         "当前请求偏好看海或海滨活动",
			SourceMemoryID: "current_request",
		})
	}

	if containsAnyConstraintText(text, []string{"老街", "市区老街", "历史街区", "步行街"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"historic_site", "city_walk", "commercial_area", "low_exertion"},
			PreferKeywords: []string{"老街", "历史街区", "步行街", "古街"},
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
			Reason:         "当前请求偏好市区老街或历史街区",
			SourceMemoryID: "current_request",
		})
	}

	if !childFriendly && containsAnyConstraintText(text, []string{"亲子", "带孩子", "带小孩"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"family", "zoo", "aquarium", "science_museum", "low_exertion"},
			AvoidTags:      []string{"high_exertion", "crowded", "remote"},
			PreferKeywords: []string{"动物园", "海洋馆", "科技馆"},
			Strength:       "soft",
			Priority:       90,
			Source:         "current_request",
			Reason:         "当前请求是亲子场景",
			SourceMemoryID: "current_request",
		})
	}

	if !avoidShopping && containsAnyConstraintText(text, []string{"购物", "商场", "买东西", "逛街"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"shopping", "commercial_area", "city_walk"},
			PreferKeywords: []string{"商场", "购物中心", "商业街", "市场"},
			Strength:       "soft",
			Priority:       80,
			Source:         "current_request",
			Reason:         "当前请求偏好购物或逛街",
			SourceMemoryID: "current_request",
		})
	}

	if containsAnyConstraintText(text, []string{"自然风景", "看风景", "公园", "湖", "森林"}) {
		constraints = append(constraints, model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"nature", "park", "lake", "forest", "outdoor"},
			PreferKeywords: []string{"公园", "湖", "森林", "自然风景"},
			Strength:       "soft",
			Priority:       80,
			Source:         "current_request",
			Reason:         "当前请求偏好自然风景",
			SourceMemoryID: "current_request",
		})
	}

	return deduplicateConstraints(normalizeConstraints(constraints))
}

func parseConstraintsFromMetadata(memory model.MemoryRecord) []model.PreferenceConstraint {
	text := strings.TrimSpace(memory.MetadataJSON)
	if text == "" {
		return nil
	}

	var payload memoryConstraintPayload
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return nil
	}

	result := make([]model.PreferenceConstraint, 0)

	for _, c := range payload.Constraints {
		c = withLongTermDefaults(c, memory)
		result = append(result, c)
	}

	excludeKeywords := append([]string{}, payload.ExcludeKeywords...)
	excludeKeywords = append(excludeKeywords, payload.NegativeKeywords...)

	preferKeywords := append([]string{}, payload.PreferKeywords...)
	preferKeywords = append(preferKeywords, payload.PositiveKeywords...)

	if len(excludeKeywords) == 0 &&
		len(preferKeywords) == 0 &&
		len(payload.AvoidTags) == 0 &&
		len(payload.PreferTags) == 0 {
		return result
	}

	domain := payload.Domain
	if domain == "" {
		domain = inferDomainFromScenario(payload.Scenario, memory.MemoryType)
	}

	strength := payload.Strength
	if strength == "" {
		strength = "hard"
	}

	reason := payload.Reason
	if reason == "" {
		reason = memory.Text
	}

	constraint := model.PreferenceConstraint{
		Domain:          domain,
		ExcludeKeywords: excludeKeywords,
		PreferKeywords:  preferKeywords,
		PreferTags:      payload.PreferTags,
		AvoidTags:       payload.AvoidTags,
		Strength:        strength,
		Priority:        payload.Priority,
		Source:          payload.Source,
		Reason:          reason,
		SourceMemoryID:  memory.ID,
	}

	result = append(result, withLongTermDefaults(constraint, memory))

	return result
}

func inferConstraintsFromText(memory model.MemoryRecord) []model.PreferenceConstraint {
	text := strings.ToLower(memory.Text)
	result := make([]model.PreferenceConstraint, 0)

	if containsAnyConstraintText(text, []string{
		"不喜欢青年旅舍",
		"不住青旅",
		"不要青旅",
		"避开青旅",
		"多人间",
		"床位房",
		"青年住宿",
	}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:          "hotel",
			ExcludeKeywords: []string{"青旅", "青年旅舍", "青年酒店", "青年旅馆", "多人间", "床位房", "hostel"},
			PreferKeywords:  []string{"地铁", "连锁", "经济型", "酒店", "速8", "布丁", "如家", "汉庭", "7天", "全季", "海友"},
			Strength:        "hard",
			Reason:          memory.Text,
			SourceMemoryID:  memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"不住民宿", "不要民宿", "不喜欢民宿"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:          "hotel",
			ExcludeKeywords: []string{"民宿", "客栈", "家庭式", "家庭旅馆", "homestay", "inn"},
			Strength:        "hard",
			Reason:          memory.Text,
			SourceMemoryID:  memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"只坐高铁", "只要高铁", "默认高铁", "优先高铁"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:          "transport",
			ExcludeKeywords: []string{"普快", "硬座", "飞机", "航班", "机场"},
			PreferKeywords:  []string{"高铁", "动车", "二等座"},
			Strength:        "hard",
			Reason:          memory.Text,
			SourceMemoryID:  memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"不坐飞机", "不要飞机", "不喜欢坐飞机", "尽量别坐飞机"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:          "transport",
			ExcludeKeywords: []string{"飞机", "航班", "机场"},
			PreferKeywords:  []string{"高铁", "动车", "火车"},
			Strength:        "hard",
			Reason:          memory.Text,
			SourceMemoryID:  memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"不坐普快", "不要普快", "不坐硬座", "不要硬座"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:          "transport",
			ExcludeKeywords: []string{"普快", "硬座"},
			PreferKeywords:  []string{"高铁", "动车", "飞机"},
			Strength:        "hard",
			Reason:          memory.Text,
			SourceMemoryID:  memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"喜欢飞机", "优先飞机", "尽量飞机"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:         "transport",
			PreferKeywords: []string{"飞机", "航班", "直达"},
			Strength:       "soft",
			Reason:         memory.Text,
			SourceMemoryID: memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"喜欢博物馆", "爱逛博物馆", "想逛博物馆", "博物馆", "展览馆", "喜欢展览", "历史人文", "喜欢历史人文"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:         "attraction",
			PreferTags:     []string{"museum", "exhibition", "art_gallery", "culture", "indoor", "low_exertion"},
			PreferKeywords: []string{"博物馆", "展览馆", "美术馆", "纪念馆", "历史", "人文", "古迹"},
			Strength:       "soft",
			Reason:         memory.Text,
			SourceMemoryID: memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"不喜欢爬山", "不要爬山", "不想爬山", "不喜欢徒步", "不要徒步"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:          "attraction",
			AvoidTags:       []string{"mountain", "hiking", "high_exertion"},
			ExcludeKeywords: []string{"爬山", "登山", "徒步", "峡谷", "山地", "山湖田园", "长城", "凤凰岭"},
			PreferTags:      []string{"low_exertion", "indoor", "culture"},
			PreferKeywords:  []string{"博物馆", "展览馆", "城市地标", "历史街区"},
			Strength:        "soft",
			Reason:          memory.Text,
			SourceMemoryID:  memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"不去远郊", "不要远郊", "不想跑太远", "景点不要太远"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:          "attraction",
			AvoidTags:       []string{"remote", "high_exertion"},
			ExcludeKeywords: []string{"远郊", "郊区", "偏远", "山", "峡谷"},
			PreferTags:      []string{"city_walk", "indoor", "low_exertion"},
			PreferKeywords:  []string{"市区", "地铁", "博物馆", "展览馆", "城市地标"},
			Strength:        "soft",
			Reason:          memory.Text,
			SourceMemoryID:  memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"轻松", "轻松一点", "慢节奏", "不想太赶", "不要特种兵", "行程轻松"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:          "route",
			ExcludeKeywords: []string{"tight_schedule", "long_transfer", "too_many_attractions"},
			PreferKeywords:  []string{"relaxed", "slow_pace", "short_transfer"},
			PreferTags:      []string{"relaxed", "slow_pace", "short_transfer"},
			AvoidTags:       []string{"tight_schedule", "long_transfer", "too_many_attractions"},
			Strength:        "soft",
			Reason:          memory.Text,
			SourceMemoryID:  memory.ID,
		}, memory))
	}

	if containsAnyConstraintText(text, []string{"特种兵", "多玩几个", "景点多一点", "行程紧凑", "不怕累"}) {
		result = append(result, withLongTermDefaults(model.PreferenceConstraint{
			Domain:         "route",
			PreferKeywords: []string{"intensive", "more_attractions", "tight_schedule"},
			PreferTags:     []string{"intensive", "more_attractions", "tight_schedule"},
			Strength:       "soft",
			Reason:         memory.Text,
			SourceMemoryID: memory.ID,
		}, memory))
	}

	return result
}

func withLongTermDefaults(constraint model.PreferenceConstraint, memory model.MemoryRecord) model.PreferenceConstraint {
	if constraint.SourceMemoryID == "" {
		constraint.SourceMemoryID = memory.ID
	}

	if constraint.Reason == "" {
		constraint.Reason = memory.Text
	}

	if constraint.Source == "" {
		constraint.Source = "long_term_memory"
	}

	if constraint.Priority <= 0 {
		constraint.Priority = 50 + memory.Importance
		if constraint.Priority > 80 {
			constraint.Priority = 80
		}
	}

	return constraint
}

func inferDomainFromScenario(scenario string, memoryType string) string {
	text := strings.ToLower(scenario + " " + memoryType)

	if strings.Contains(text, "hotel") || strings.Contains(text, "住宿") {
		return "hotel"
	}

	if strings.Contains(text, "transport") || strings.Contains(text, "交通") {
		return "transport"
	}

	if strings.Contains(text, "attraction") || strings.Contains(text, "景点") {
		return "attraction"
	}

	if strings.Contains(text, "food") || strings.Contains(text, "美食") {
		return "food"
	}

	if strings.Contains(text, "route") || strings.Contains(text, "路线") {
		return "route"
	}

	return "general"
}

func normalizeConstraints(constraints []model.PreferenceConstraint) []model.PreferenceConstraint {
	result := make([]model.PreferenceConstraint, 0, len(constraints))

	for _, c := range constraints {
		c.Domain = strings.TrimSpace(strings.ToLower(c.Domain))
		if c.Domain == "" {
			c.Domain = "general"
		}

		c = enrichConstraintTagsFromKeywords(c)

		c.Strength = strings.TrimSpace(strings.ToLower(c.Strength))
		if c.Strength == "" {
			c.Strength = "hard"
		}

		c.Source = strings.TrimSpace(strings.ToLower(c.Source))
		if c.Source == "" {
			c.Source = "long_term_memory"
		}

		c.ExcludeKeywords = normalizeKeywordList(c.ExcludeKeywords)
		c.PreferKeywords = normalizeKeywordList(c.PreferKeywords)
		c.PreferTags = normalizeKeywordList(c.PreferTags)
		c.AvoidTags = normalizeKeywordList(c.AvoidTags)

		if len(c.ExcludeKeywords) == 0 &&
			len(c.PreferKeywords) == 0 &&
			len(c.PreferTags) == 0 &&
			len(c.AvoidTags) == 0 {
			continue
		}

		result = append(result, c)
	}

	return result
}

func enrichConstraintTagsFromKeywords(c model.PreferenceConstraint) model.PreferenceConstraint {
	text := strings.ToLower(strings.Join(append(append([]string{}, c.PreferKeywords...), c.ExcludeKeywords...), " "))

	if c.Domain == "attraction" || c.Domain == "general" {
		if len(c.PreferTags) == 0 {
			if containsAnyConstraintText(text, []string{"博物馆", "展览馆", "美术馆", "纪念馆", "历史", "人文", "古迹"}) {
				c.PreferTags = append(c.PreferTags, "museum", "exhibition", "art_gallery", "memorial", "culture", "indoor", "low_exertion")
			}

			if containsAnyConstraintText(text, []string{"公园", "自然", "湖", "森林"}) {
				c.PreferTags = append(c.PreferTags, "nature", "park", "outdoor")
			}

			if containsAnyConstraintText(text, []string{"地标", "夜景", "观景"}) {
				c.PreferTags = append(c.PreferTags, "landmark", "night_view", "city_walk")
			}

			if containsAnyConstraintText(text, []string{"看海", "海边", "海滩", "沙滩", "海滨", "海岸", "滨海", "观海"}) {
				c.PreferTags = append(c.PreferTags, "sea", "beach", "waterfront", "coast", "low_exertion")
			}
		}

		if len(c.AvoidTags) == 0 {
			if containsAnyConstraintText(text, []string{"爬山", "登山", "徒步", "长城", "山", "峡谷", "凤凰岭"}) {
				c.AvoidTags = append(c.AvoidTags, "mountain", "hiking", "high_exertion")
			}

			if containsAnyConstraintText(text, []string{"远郊", "郊区", "偏远"}) {
				c.AvoidTags = append(c.AvoidTags, "remote")
			}
		}
	}

	if c.Domain == "route" || c.Domain == "general" {
		if len(c.PreferTags) == 0 && containsAnyConstraintText(text, []string{"relaxed", "slow_pace", "short_transfer"}) {
			c.PreferTags = append(c.PreferTags, "relaxed", "slow_pace", "short_transfer")
		}

		if len(c.AvoidTags) == 0 && containsAnyConstraintText(text, []string{"tight_schedule", "long_transfer", "too_many_attractions"}) {
			c.AvoidTags = append(c.AvoidTags, "tight_schedule", "long_transfer", "too_many_attractions")
		}
	}

	if c.Domain == "food" || c.Domain == "general" {
		if len(c.PreferTags) == 0 && containsAnyConstraintText(text, []string{"小吃", "美食", "夜市", "本地美食"}) {
			c.PreferTags = append(c.PreferTags, "food", "local_food", "snack_street")
		}
	}

	return c
}

func normalizeKeywordList(keywords []string) []string {
	result := make([]string, 0, len(keywords))
	seen := map[string]bool{}

	for _, keyword := range keywords {
		keyword = strings.TrimSpace(strings.ToLower(keyword))
		if keyword == "" {
			continue
		}

		if seen[keyword] {
			continue
		}

		seen[keyword] = true
		result = append(result, keyword)
	}

	return result
}

func deduplicateConstraints(constraints []model.PreferenceConstraint) []model.PreferenceConstraint {
	result := make([]model.PreferenceConstraint, 0, len(constraints))
	seen := map[string]bool{}

	for _, c := range constraints {
		key := strings.Join([]string{
			c.Domain,
			c.Strength,
			c.Source,
			strings.Join(c.ExcludeKeywords, ","),
			strings.Join(c.PreferKeywords, ","),
			strings.Join(c.AvoidTags, ","),
			strings.Join(c.PreferTags, ","),
		}, "|")

		if seen[key] {
			continue
		}

		seen[key] = true
		result = append(result, c)
	}

	return result
}

func containsAnyConstraintText(text string, keywords []string) bool {
	text = strings.ToLower(text)

	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}

		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}

func hasNegativeAccommodationIntent(text string, accommodationTerms []string) bool {
	negativePrefixes := []string{"不要", "不住", "不想住", "避开", "别推荐", "不推荐", "不考虑", "不喜欢"}

	for _, term := range accommodationTerms {
		for _, prefix := range negativePrefixes {
			if strings.Contains(text, prefix+term) {
				return true
			}
		}
	}

	return false
}

func hasPositiveAccommodationIntent(text string, accommodationTerms []string) bool {
	positivePrefixes := []string{"想体验", "想住", "希望", "特色", "海边", "住", "体验"}

	for _, term := range accommodationTerms {
		for _, prefix := range positivePrefixes {
			if strings.Contains(text, prefix+term) || strings.Contains(text, prefix+"的"+term) {
				return true
			}
		}
	}

	return false
}

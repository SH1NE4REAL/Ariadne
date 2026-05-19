package tools

import (
	"strings"

	"ariadne/internal/model"
)

func BuildPOIProfile(attraction model.Attraction) model.POIProfile {
	semanticDescription := attraction.Description
	if strings.Contains(semanticDescription, "搜索关键词") {
		semanticDescription = ""
	}

	text := strings.ToLower(strings.Join([]string{
		attraction.Name,
		attraction.Category,
		semanticDescription,
		attraction.Address,
	}, " "))

	profile := model.POIProfile{
		Name:        attraction.Name,
		Category:    attraction.Category,
		Description: attraction.Description,
		Address:     attraction.Address,
		Attraction:  attraction,
		Tags:        make([]string, 0),
	}

	if isInvalidPOI(text, attraction.Name, attraction.Category) {
		profile.Invalid = true
		profile.InvalidReason = "not_core_attraction"
		return profile
	}

	profile.Tags = inferPOITags(text)
	return profile
}

func isInvalidPOI(text string, name string, category string) bool {
	invalidKeywords := []string{
		"东门", "西门", "南门", "北门",
		"东北门", "东南门", "西南门", "西北门",
		"入口", "出口", "出入口",
		"停车场", "售票处", "游客中心",
		"办公室", "服务中心",
		"通行设施", "门/出入口",
		"室内及附属设施",
		"教育学校:教育学校附属",
		"历史学系", "人文楼", "人文学院", "办公楼",
		"民宿", "酒店", "宾馆", "公寓",
		"彩票",
		"住宅区", "小区", "社区", "居民楼",
		"公交站", "公交车站", "地铁站出入口",
		"口腔", "医院", "诊所", "药房",
		"村委会", "居委会", "政府机关", "派出所", "办事处",
		"照明工程", "灯具用品", "灯饰",
		"安置房", "农贸市场",
		"建材", "家具", "道路名", "行政地名",
		"学院", "管理咨询中心",
		"武术俱乐部", "俱乐部", "烟酒",
	}

	invalidCategories := []string{
		"酒店宾馆",
		"生活服务",
		"汽车:停车场",
		"教育学校",
		"室内及附属设施",
		"房地产",
		"住宅区",
		"医疗保健",
		"政府机构",
		"交通设施",
		"道路",
		"行政地名",
		"建材",
		"家具",
	}

	if containsAnyText(text, invalidKeywords) {
		return true
	}

	if containsAnyText(category, invalidCategories) {
		return true
	}

	if containsAnyText(text, []string{"失恋博物馆", "成人", "情侣主题", "酒吧", "夜店"}) {
		return true
	}

	name = strings.TrimSpace(name)

	if strings.Contains(name, "-") && containsAnyText(name, []string{
		"东门", "西门", "南门", "北门",
		"东北门", "东南门", "西南门", "西北门",
	}) {
		return true
	}

	return false
}

func inferPOITags(text string) []string {
	tags := make([]string, 0)

	if containsAnyText(text, []string{"博物馆"}) {
		tags = append(tags, "museum", "culture", "indoor", "low_exertion")
	}

	if containsAnyText(text, []string{"展览馆", "展览中心", "展厅"}) {
		tags = append(tags, "exhibition", "culture", "indoor", "low_exertion")
	}

	if containsAnyText(text, []string{"美术馆", "艺术馆"}) {
		tags = append(tags, "art_gallery", "culture", "indoor", "low_exertion")
	}

	if containsAnyText(text, []string{"纪念馆", "档案馆"}) {
		tags = append(tags, "memorial", "culture", "indoor", "low_exertion")
	}

	if containsAnyText(text, []string{"历史古迹", "遗址", "古城", "古镇", "胡同", "故宫", "圆明园"}) {
		tags = append(tags, "historic_site", "culture", "city_walk")
	}

	if containsAnyText(text, []string{"公园", "园林", "植物园", "花园"}) {
		tags = append(tags, "park", "nature", "outdoor", "medium_exertion")
	}

	if containsAnyText(text, []string{"山", "长城", "峡谷", "森林", "徒步", "登山", "爬山", "凤凰岭"}) {
		tags = append(tags, "mountain", "hiking", "outdoor", "high_exertion")
	}

	if containsAnyText(text, []string{"动物园"}) {
		tags = append(tags, "zoo", "family", "outdoor")
	}

	marketOrPetText := containsAnyText(text, []string{"花鸟鱼虫", "宠物市场", "花鸟市场", "鱼虫市场", "市场"})

	if containsAnyText(text, []string{"海洋馆", "水族馆"}) && !marketOrPetText {
		tags = append(tags, "aquarium", "family", "indoor")
	}

	if containsAnyText(text, []string{"科技馆", "科学中心"}) {
		tags = append(tags, "science_museum", "family", "culture", "indoor")
	}

	if containsAnyText(text, []string{"儿童", "亲子", "科普"}) {
		tags = append(tags, "family", "child_friendly")
	}

	if containsAnyText(text, []string{"商场", "购物中心", "商业街", "市场"}) {
		tags = append(tags, "shopping", "commercial_area", "city_walk")
	}

	if containsAnyText(text, []string{"小吃", "夜市"}) {
		tags = append(tags, "food", "local_food", "snack_street")
	}

	if containsAnyText(text, []string{"美食", "餐厅", "饭店", "茶馆", "茶社", "咖啡"}) {
		tags = append(tags, "food", "local_food")
	}

	if containsAnyText(text, []string{"海边", "海滩", "沙滩", "海滨", "海岸", "滨海", "观海", "码头", "海湾", "海堤", "栈道"}) {
		tags = append(tags, "sea", "beach", "waterfront", "coast", "low_exertion")
	}

	if containsAnyText(text, []string{"夜景", "观景", "电视塔", "高楼", "天台"}) {
		tags = append(tags, "night_view", "landmark")
	}

	if containsAnyText(text, []string{"酒吧", "夜店", "夜生活"}) {
		tags = append(tags, "bar", "nightlife", "adult_or_couple")
	}

	if containsAnyText(text, []string{"失恋博物馆", "情侣", "成人"}) {
		tags = append(tags, "adult_or_couple")
	}

	if containsAnyText(text, []string{"网红", "打卡"}) {
		tags = append(tags, "influencer_spot", "crowded")
	}

	if len(tags) == 0 {
		tags = append(tags, "general_attraction")
	}

	return uniqueStringList(tags)
}

func containsAnyText(text string, keywords []string) bool {
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

func uniqueStringList(items []string) []string {
	result := make([]string, 0, len(items))
	seen := map[string]bool{}

	for _, item := range items {
		item = strings.TrimSpace(strings.ToLower(item))
		if item == "" || seen[item] {
			continue
		}

		seen[item] = true
		result = append(result, item)
	}

	return result
}

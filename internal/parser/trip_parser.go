package parser

import (
	"ariadne/internal/model"
	"regexp"
	"strconv"
	"strings"
)

func ParseTripRequest(rawInput string) model.TripRequest {
	request := model.TripRequest{
		RawInput: rawInput,
	}

	request.Origin = parseOrigin(rawInput)
	request.Destination = parseDestination(rawInput)
	request.Days = parseDays(rawInput)
	request.Budget = parseBudget(rawInput)
	request.Preference = parsePreference(rawInput)
	request.TransportPreference = parseTransportPreference(rawInput)

	return request
}

func parseOrigin(text string) string {
	re := regexp.MustCompile(`从(.+?)去`)
	match := re.FindStringSubmatch(text)

	if len(match) >= 2 {
		return match[1]
	}

	return ""
}

func parseDestination(text string) string {
	re := regexp.MustCompile(`去(.+?)(玩|旅游|出发|，|,|。|$)`)
	match := re.FindStringSubmatch(text)

	if len(match) >= 2 {
		return match[1]
	}

	return ""
}

func parseDays(text string) int {
	if strings.Contains(text, "三天") {
		return 3
	}
	if strings.Contains(text, "两天") || strings.Contains(text, "二天") {
		return 2
	}
	if strings.Contains(text, "一天") {
		return 1
	}

	re := regexp.MustCompile(`(\d+)天`)
	match := re.FindStringSubmatch(text)

	if len(match) >= 2 {
		days, err := strconv.Atoi(match[1])
		if err == nil {
			return days
		}
	}

	return 0
}

func parseBudget(text string) int {
	re := regexp.MustCompile(`预算(\d+)`)
	match := re.FindStringSubmatch(text)

	if len(match) >= 2 {
		budget, err := strconv.Atoi(match[1])
		if err == nil {
			return budget
		}
	}

	return 0
}

func parsePreference(text string) string {
	if strings.Contains(text, "轻松") || strings.Contains(text, "不想太赶") {
		return "轻松"
	}
	if strings.Contains(text, "省钱") || strings.Contains(text, "便宜") {
		return "省钱"
	}
	if strings.Contains(text, "美食") {
		return "美食"
	}
	if strings.Contains(text, "拍照") {
		return "拍照"
	}

	return ""
}
func parseTransportPreference(text string) string {
	if strings.Contains(text, "高铁") || strings.Contains(text, "动车") || strings.Contains(text, "火车") {
		return "高铁"
	}

	if strings.Contains(text, "飞机") || strings.Contains(text, "机票") || strings.Contains(text, "航班") {
		return "飞机"
	}

	if strings.Contains(text, "自驾") || strings.Contains(text, "开车") {
		return "自驾"
	}

	return ""
}
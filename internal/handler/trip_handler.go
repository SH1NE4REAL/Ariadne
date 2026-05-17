package handler

import (
	"encoding/json"
	"net/http"

	"ariadne/internal/model"
	"ariadne/internal/service"
)

func PlanTripHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, model.ApiResponse{
			Code:    http.StatusMethodNotAllowed,
			Message: "只支持 POST 请求",
			Data:    nil,
		})
		return
	}

	var req model.TripPlanRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "请求 JSON 格式错误",
			Data:    nil,
		})
		return
	}

	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, model.ApiResponse{
			Code:    http.StatusBadRequest,
			Message: "message 不能为空",
			Data:    nil,
		})
		return
	}

	agentResult := service.RunTripAgent(req.Message)

	if agentResult.NeedClarification {
		writeJSON(w, http.StatusOK, model.ApiResponse{
			Code:    1001,
			Message: "需要补充旅行信息",
			Data:    agentResult.Clarification,
		})
		return
	}

	writeJSON(w, http.StatusOK, model.ApiResponse{
		Code:    0,
		Message: "success",
		Data:    agentResult.FinalPlan,
	})
	}

	func writeJSON(w http.ResponseWriter, statusCode int, response model.ApiResponse) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
}
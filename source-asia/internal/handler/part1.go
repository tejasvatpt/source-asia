package handler

import (
	"encoding/json"
	"net/http"

	"github.com/tejasvatpt/source-asia/internal/models"
	"github.com/tejasvatpt/source-asia/internal/ratelimit"
)

type Part1Handler struct {
	limiter *ratelimit.Limiter
}

func NewPart1Handler(l *ratelimit.Limiter) *Part1Handler {
	return &Part1Handler{limiter: l}
}

func (h *Part1Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use POST")
		return
	}

	var body models.RequestBody

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON: "+err.Error())
		return
	}

	if body.UserID == "" {
		writeError(w, http.StatusBadRequest, "missing_user_id", "user_id is required and must be non-empty")
		return
	}

	if body.Payload == nil {
		writeError(w, http.StatusBadRequest, "missing_payload", "payload is required")
		return
	}

	if !h.limiter.Allow(body.UserID) {
		writeError(w, http.StatusTooManyRequests, "rate_limit_exceeded",
			"you have exceeded the limit of 5 requests per 60-second window")
		return
	}

	writeJSON(w, http.StatusCreated, models.RequestResponse{
		Message:  "request accepted",
		UserID:   body.UserID,
		Accepted: true,
	})
}

func (h *Part1Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "use GET")
		return
	}

	userIDs := h.limiter.AllUserIDs()
	stats := make([]models.UserStats, 0, len(userIDs))

	for _, uid := range userIDs {
		accepted, rejected := h.limiter.Stats(uid)
		stats = append(stats, models.UserStats{
			UserID:           uid,
			AcceptedInWindow: accepted,
			RejectedTotal:    rejected,
		})
	}

	writeJSON(w, http.StatusOK, models.StatsResponse{Users: stats})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, errKey, message string) {
	writeJSON(w, status, models.ErrorResponse{
		Error:   errKey,
		Message: message,
	})
}

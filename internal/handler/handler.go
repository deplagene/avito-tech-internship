package handler

import (
	"deplagene/avito-tech-internship/api"
	"deplagene/avito-tech-internship/types"
	"encoding/json"
	"log/slog"
	"net/http"
)

// Handler реализует интерфейс api.ServerInterface.
type Handler struct {
	teamService types.TeamService
	userService types.UserService
	prService   types.PullRequestService
	logger      *slog.Logger
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(
	teamService types.TeamService,
	userService types.UserService,
	prService types.PullRequestService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		teamService: teamService,
		userService: userService,
		prService:   prService,
		logger:      logger,
	}
}

// sendError отправляет стандартизированный ответ об ошибке.
func (h *Handler) sendError(w http.ResponseWriter, r *http.Request, code api.ErrorResponseErrorCode, message string, httpStatus int) {
	h.logger.Error("API Error", "code", code, "message", message, "status", httpStatus, "path", r.URL.Path)
	resp := api.ErrorResponse{
		Error: struct {
			Code    api.ErrorResponseErrorCode "json:\"code\""
			Message string                     "json:\"message\""
		}{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}

// PostPullRequestCreate создает PR и автоматически назначает до 2 ревьюверов из команды автора
func (h *Handler) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestCreateJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.sendError(w, r, "BAD_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	pr := api.PullRequest{
		PullRequestId:   body.PullRequestId,
		PullRequestName: body.PullRequestName,
		AuthorId:        body.AuthorId,
	}

	createdPR, err := h.prService.CreatePullRequest(r.Context(), pr)
	if err != nil {
		// TODO: Map service errors to API error codes
		h.sendError(w, r, "INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Pr api.PullRequest `json:"pr"`
	}{Pr: *createdPR})
}

// PostPullRequestMerge помечает PR как MERGED (идемпотентная операция)
func (h *Handler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestMergeJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.sendError(w, r, "BAD_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	mergedPR, err := h.prService.MergePullRequest(r.Context(), body.PullRequestId)
	if err != nil {
		// TODO: Map service errors to API error codes
		h.sendError(w, r, "INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Pr api.PullRequest `json:"pr"`
	}{Pr: *mergedPR})
}

// PostPullRequestReassign переназначает конкретного ревьювера на другого из его команды
func (h *Handler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestReassignJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.sendError(w, r, "BAD_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	reassignedPR, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		// TODO: Map service errors to API error codes
		h.sendError(w, r, "INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Pr         api.PullRequest `json:"pr"`
		ReplacedBy string          `json:"replaced_by"`
	}{Pr: *reassignedPR, ReplacedBy: newReviewerID})
}

// PostTeamAdd создает команду с участниками (создаёт/обновляет пользователей)
func (h *Handler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var body api.PostTeamAddJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.sendError(w, r, "BAD_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	createdTeam, err := h.teamService.CreateTeam(r.Context(), body)
	if err != nil {
		// TODO: Map service errors to API error codes
		h.sendError(w, r, "INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Team api.Team `json:"team"`
	}{Team: *createdTeam})
}

// GetTeamGet получает команду с участниками
func (h *Handler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	team, err := h.teamService.GetTeam(r.Context(), params.TeamName)
	if err != nil {
		// TODO: Map service errors to API error codes
		h.sendError(w, r, "INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(team)
}

// GetUsersGetReview получает PR'ы, где пользователь назначен ревьювером
func (h *Handler) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	prs, err := h.prService.GetPullRequestsByReviewer(r.Context(), params.UserId)
	if err != nil {
		// TODO: Map service errors to API error codes
		h.sendError(w, r, "INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		UserId       string                 `json:"user_id"`
		PullRequests []api.PullRequestShort `json:"pull_requests"`
	}{UserId: params.UserId, PullRequests: prs})
}

// PostUsersSetIsActive устанавливает флаг активности пользователя
func (h *Handler) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body api.PostUsersSetIsActiveJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.sendError(w, r, "BAD_REQUEST", "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedUser, err := h.userService.SetUserIsActive(r.Context(), body.UserId, body.IsActive)
	if err != nil {
		// TODO: Map service errors to API error codes
		h.sendError(w, r, "INTERNAL_SERVER_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		User api.User `json:"user"`
	}{User: *updatedUser})
}

// GetHealth implements the /health endpoint.
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Проверка соответствия интерфейсу во время компиляции
var _ api.ServerInterface = (*Handler)(nil)
var _ api.ServerInterface = (*Handler)(nil) // This line is a duplicate, will be fixed in the next step


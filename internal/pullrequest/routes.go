package pullrequest

import (
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/types"
	"deplagene/avito-tech-internship/utils"
	"log/slog"
	"net/http"
)

type Handler struct {
	teamService types.TeamService
	userService types.UserService
	prService   types.PullRequestService
	logger      *slog.Logger
}

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

	if err := utils.WriteJson(w, httpStatus, resp); err != nil {
		utils.WriteError(w, h.logger, httpStatus, err)
	}
}

// PostPullRequestCreate создает PR и автоматически назначает до 2 ревьюверов из команды автора
func (h *Handler) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestCreateJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
	}

	pr := api.PullRequest{
		PullRequestId:   body.PullRequestId,
		PullRequestName: body.PullRequestName,
		AuthorId:        body.AuthorId,
	}

	createdPR, err := h.prService.CreatePullRequest(r.Context(), pr)
	if err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}

	if err := utils.WriteJson(w, http.StatusCreated, map[string]string{"pull_request_id": createdPR.PullRequestId}); err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// PostPullRequestMerge помечает PR как MERGED (идемпотентная операция)
func (h *Handler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestMergeJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
	}

	mergedPR, err := h.prService.MergePullRequest(r.Context(), body.PullRequestId)
	if err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}

	if err := utils.WriteJson(w, http.StatusOK, map[string]string{"pull_request_id": mergedPR.PullRequestId}); err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// PostPullRequestReassign переназначает конкретного ревьювера на другого из его команды
func (h *Handler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestReassignJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
	}

	reassignedPR, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}

	if err := utils.WriteJson(w, http.StatusOK, map[string]string{"pull_request_id": reassignedPR.PullRequestId, "new_reviewer_id": newReviewerID}); err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// PostTeamAdd создает команду с участниками (создаёт/обновляет пользователей)
func (h *Handler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var body api.PostTeamAddJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
	}

	createdTeam, err := h.teamService.CreateTeam(r.Context(), body)
	if err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}

	if err := utils.WriteJson(w, http.StatusCreated, map[string]string{"team_name": createdTeam.TeamName}); err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// GetTeamGet получает команду с участниками
func (h *Handler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	team, err := h.teamService.GetTeam(r.Context(), params.TeamName)
	if err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}

	if err := utils.WriteJson(w, http.StatusOK, map[string]string{"team_name": team.TeamName}); err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// GetUsersGetReview получает PR'ы, где пользователь назначен ревьювером
func (h *Handler) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	prs, err := h.prService.GetPullRequestsByReviewer(r.Context(), params.UserId)
	if err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}

	if err := utils.WriteJson(w, http.StatusOK, prs); err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// PostUsersSetIsActive устанавливает флаг активности пользователя
func (h *Handler) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body api.PostUsersSetIsActiveJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
	}

	updatedUser, err := h.userService.SetUserIsActive(r.Context(), body.UserId, body.IsActive)
	if err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}

	if err := utils.WriteJson(w, http.StatusOK, map[string]string{"user_id": updatedUser.UserId}); err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// GetHealth проверяет работоспособность сервиса
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	if err := utils.WriteJson(w, http.StatusOK, map[string]string{"status": "ok"}); err != nil {
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// Проверка соответствия интерфейсу во время компиляции
var _ api.ServerInterface = (*Handler)(nil)

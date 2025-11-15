package pullrequest

import (
	"deplagene/avito-tech-internship/cmd/api"
	"deplagene/avito-tech-internship/types"
	"deplagene/avito-tech-internship/utils"
	"errors"
	"fmt"
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

// handleError отправляет стандартизированный ответ об ошибке.
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	h.logger.Info("handleError received error", "error", err.Error(), "type", fmt.Sprintf("%T", err))
	var code api.ErrorResponseErrorCode
	var message string
	var httpStatus int

	switch {
	case errors.Is(err, types.ErrNotFound):
		code = api.NOTFOUND
		message = "resource not found"
		httpStatus = http.StatusNotFound
	case errors.Is(err, types.ErrAlreadyExists):
		code = api.TEAMEXISTS
		message = "resource already exists"
		httpStatus = http.StatusConflict
	case errors.Is(err, types.ErrPRMerged):
		code = api.PRMERGED
		message = "cannot reassign on merged PR"
		httpStatus = http.StatusConflict
	case errors.Is(err, types.ErrNotAssigned):
		code = api.NOTASSIGNED
		message = "reviewer is not assigned to this PR"
		httpStatus = http.StatusConflict
	case errors.Is(err, types.ErrNoCandidate):
		code = api.NOCANDIDATE
		message = "no active replacement candidate in team"
		httpStatus = http.StatusConflict
	default:
		h.logger.Error("Internal Server Error", "error", err, "path", r.URL.Path)
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
		return
	}

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
		h.logger.Error("Failed to write error response", "error", err, "path", r.URL.Path)
		utils.WriteError(w, h.logger, http.StatusInternalServerError, err)
	}
}

// PostPullRequestCreate создает PR и автоматически назначает до 2 ревьюверов из команды автора
func (h *Handler) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestCreateJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	pr := api.PullRequest{
		PullRequestId:   body.PullRequestId,
		PullRequestName: body.PullRequestName,
		AuthorId:        body.AuthorId,
	}

	createdPR, err := h.prService.CreatePullRequest(r.Context(), pr)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	response := struct {
		PR *api.PullRequest `json:"pr"`
	}{
		PR: createdPR,
	}

	if err := utils.WriteJson(w, http.StatusCreated, response); err != nil {
		h.handleError(w, r, err)
	}
}

// PostPullRequestMerge помечает PR как MERGED (идемпотентная операция)
func (h *Handler) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestMergeJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	mergedPR, err := h.prService.MergePullRequest(r.Context(), body.PullRequestId)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	response := struct {
		PR *api.PullRequest `json:"pr"`
	}{
		PR: mergedPR,
	}

	if err := utils.WriteJson(w, http.StatusOK, response); err != nil {
		h.handleError(w, r, err)
	}
}

// PostPullRequestReassign переназначает конкретного ревьювера на другого из его команды
func (h *Handler) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var body api.PostPullRequestReassignJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	reassignedPR, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), body.PullRequestId, body.OldUserId)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	response := struct {
		PR         *api.PullRequest `json:"pr"`
		ReplacedBy string           `json:"replaced_by"`
	}{
		PR:         reassignedPR,
		ReplacedBy: newReviewerID,
	}

	if err := utils.WriteJson(w, http.StatusOK, response); err != nil {
		h.handleError(w, r, err)
	}
}

// PostTeamAdd создает команду с участниками (создаёт/обновляет пользователей)
func (h *Handler) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var body api.PostTeamAddJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	createdTeam, err := h.teamService.CreateTeam(r.Context(), body)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	response := struct {
		Team *api.Team `json:"team"`
	}{
		Team: createdTeam,
	}

	if err := utils.WriteJson(w, http.StatusCreated, response); err != nil {
		h.handleError(w, r, err)
	}
}

// GetTeamGet получает команду с участниками
func (h *Handler) GetTeamGet(w http.ResponseWriter, r *http.Request, params api.GetTeamGetParams) {
	team, err := h.teamService.GetTeam(r.Context(), params.TeamName)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	if err := utils.WriteJson(w, http.StatusOK, team); err != nil {
		h.handleError(w, r, err)
	}
}

// GetUsersGetReview получает PR'ы, где пользователь назначен ревьювером
func (h *Handler) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params api.GetUsersGetReviewParams) {
	prs, err := h.prService.GetPullRequestsByReviewer(r.Context(), params.UserId)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	response := struct {
		UserID       string                 `json:"user_id"`
		PullRequests []api.PullRequestShort `json:"pull_requests"`
	}{
		UserID:       params.UserId,
		PullRequests: prs,
	}

	if err := utils.WriteJson(w, http.StatusOK, response); err != nil {
		h.handleError(w, r, err)
	}
}

// PostUsersSetIsActive устанавливает флаг активности пользователя
func (h *Handler) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body api.PostUsersSetIsActiveJSONRequestBody
	if err := utils.ParseJson(r, &body); err != nil {
		utils.WriteError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	updatedUser, err := h.userService.SetUserIsActive(r.Context(), body.UserId, body.IsActive)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	response := struct {
		User *api.User `json:"user"`
	}{
		User: updatedUser,
	}

	if err := utils.WriteJson(w, http.StatusOK, response); err != nil {
		h.handleError(w, r, err)
	}
}

// GetHealth проверяет работоспособность сервиса
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	if err := utils.WriteJson(w, http.StatusOK, map[string]string{"status": "ok"}); err != nil {
		h.handleError(w, r, err)
	}
}

// Проверка соответствия интерфейсу во время компиляции
var _ api.ServerInterface = (*Handler)(nil)

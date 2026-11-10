package receiver_transport_http

import (
	"errors"
	"fmt"
	"net/http"
	core_errors "notification/internal/core/errors"
	core_logger "notification/internal/core/logger"
	core_http_request "notification/internal/core/transport/http/request"
	core_http_response "notification/internal/core/transport/http/response"

	"go.uber.org/zap"
)

type AddTaskRequest struct {
	Expiration int64  `json:"expiration"`
	Phone      string `json:"phone" validate:"required,numeric,min=7,max=15"`
}

type TaskResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message,omitempty"`
}

// CreateUser 	godoc
// @Summary 	сощдать пользователя
// @Description создать нового польщователя в системе
// @Tags 		users
// @Accept 		json
// @Produce 	json
// @Param 		request body CreateUserRequest true "Create user тело запроса"
// @Success 	201 {object} CreateUserResponse "успешно созданфй пользователь"
// @Failure 	400 {object} core_http_response.ErrorResponse "Bad request"
// @Failure 	500 {object} core_http_response.ErrorResponse "Internal server error"
// @Router 		/users [post]
func (h *ReceiverHTTPHandler) Add(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_response.NewHttpResponseHandler(log, rw)

	var req AddTaskRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		log.Warn("invalid task request", zap.Error(err))
		resp.JSONResponse(TaskResponse{Error: 1, Message: "Некорректный запрос"}, http.StatusOK)
		return
	}

	if err := h.receiverService.Add(ctx, req); err != nil {
		log.Error("failed to enqueue sms task",
			zap.String("phone", req.Phone),
			zap.Error(err),
		)

		msg := "Не удалось отправить SMS"
		if errors.Is(err, core_errors.ErrTaskExpired) {
			msg = "Время задачи истекло"
		}
		resp.JSONResponse(TaskResponse{Error: 1, Message: msg}, http.StatusOK)
		return
	}

	resp.JSONResponse(TaskResponse{Error: 0, Message: fmt.Sprintf("код отправлен на номер: %s", req.Phone)}, http.StatusOK)
}

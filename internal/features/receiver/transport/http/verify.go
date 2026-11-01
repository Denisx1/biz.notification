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

type VerifyTaskRequest struct {
	Phone string `json:"phone" validate:"required,numeric,min=7,max=15"`
	Code  string `json:"code" validate:"required,len=4"`
}

func (h *ReceiverHTTPHandler) Verify(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	resp := core_http_response.NewHttpResponseHandler(log, rw)

	var req VerifyTaskRequest
	if err := core_http_request.DecodeAndValidateRequest(r, &req); err != nil {
		log.Warn("invalid task request", zap.Error(err))
		resp.JSONResponse(TaskResponse{Error: 1, Message: "Некорректный запрос"}, http.StatusOK)
		return
	}

	if err := h.receiverService.Verify(ctx, req.Phone, req.Code); err != nil {
		log.Error("failed to verify sms code",
			zap.Error(err),
		)

		msg := "Не удалось верифицировать SMS"
		if errors.Is(err, core_errors.ErrTaskExpired) {
			msg = "Время задачи истекло"
		}
		resp.JSONResponse(TaskResponse{Error: 1, Message: msg}, http.StatusOK)
		return
	}

	resp.JSONResponse(TaskResponse{Error: 0, Message: fmt.Sprintf("код верифицирован для номера: %s", req.Phone)}, http.StatusOK)
}

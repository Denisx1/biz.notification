package receiver_transport_http

import (
	"context"
	"net/http"
	core_http_server "notification/internal/core/transport/http/server"
)

type ReceiverHTTPHandler struct {
	receiverService ReceiverService
}
type ReceiverService interface {
	Add(ctx context.Context, receiver AddTaskRequest) error
	Verify(ctx context.Context, phone, code string) error
}

func NewReceiverHTTPHandler(receiverService ReceiverService) *ReceiverHTTPHandler {
	return &ReceiverHTTPHandler{
		receiverService: receiverService,
	}
}

func (h *ReceiverHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/add",
			Handler: h.Add,
		},
		{
			Method:  http.MethodPost,
			Path:    "/verify",
			Handler: h.Verify,
		},
	}
}

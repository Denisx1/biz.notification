package sender_transport_nats

import (
	"context"
	"encoding/json"
	"fmt"
	"notification/internal/core/domain"
	"notification/internal/core/transport/nats"
)

type SenderService interface {
	SendSMS(sms domain.SMSMessage) (string, error)
}
type NatsHandler struct {
	receiverService SenderService
}

func NewNatsHandler(s SenderService) *NatsHandler {
	return &NatsHandler{receiverService: s}
}

// Handle реализует интерфейс Handler для вашего NATS-клиента
func (h *NatsHandler) Handle(ctx context.Context, msg nats.Message) error {
	var m domain.SMSMessage
	if err := json.Unmarshal(msg.Data, &m); err != nil {
		return fmt.Errorf("unmarshal sms: %w", err)
	}
	h.receiverService.SendSMS(m)
	fmt.Printf("got sms: phone=%s code=%s\n", m.Phone, m.Code)
	return nil
}

package sender_service

import "context"

type ReceiverRepository interface {
	MarkSent(ctx context.Context, id int64) error
}

type SenderService struct {
	receiverRepo ReceiverRepository
	baseSMS      string
	APIKey       string
}

func NewSenderService(receiverRepo ReceiverRepository, baseSMS string, apiKey string) *SenderService {
	return &SenderService{
		receiverRepo: receiverRepo,
		baseSMS:      baseSMS,
		APIKey:       apiKey,
	}
}

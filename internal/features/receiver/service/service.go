package receiver_service

import (
	"context"
	"notification/internal/core/domain"
	core_postgres_pool "notification/internal/core/repository/postgres/pool"
)

type ReceiverRepository interface {
	LockPhone(ctx context.Context, phone string) error
	RecentCounts(ctx context.Context, phone, ip string) (Counts, error)
	Create(ctx context.Context, t *domain.SMSTask) (int64, error)
	MarkSent(ctx context.Context, id int64) error
	Get(ctx context.Context, phone string) (domain.SMSVerification, error)
	MarkFailed(ctx context.Context, id int64, reason string) error
	MarkExpired(ctx context.Context, id int64) error
	TakeAttempt(ctx context.Context, phone string) (Attempt, error)
	MarkVerified(ctx context.Context, id int64) error
	MarkError(ctx context.Context, id int64, er error) error
}

type TaskPublisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

type SMSSender interface {
	Send(ctx context.Context, phone, text string) error
}

type ReceiverService struct {
	repo      ReceiverRepository
	txManager core_postgres_pool.TxManager
	publisher TaskPublisher
}

func NewReceiverService(
	repo ReceiverRepository,
	txManager core_postgres_pool.TxManager,
	publisher TaskPublisher,
) *ReceiverService {
	return &ReceiverService{
		repo:      repo,
		txManager: txManager,
		publisher: publisher,
	}
}

// Add: вызывает HTTP. Быстро: запись в базу + публикация.

// func (s *ReceiverService) Send(ctx context.Context, task domain.SMSTask) error {
// 	if task.Expired() {
// 		return s.repo.MarkExpired(ctx, task.ID)
// 	}
// 	if err := s.sender.Send(ctx, task.Phone, task.Text); err != nil {
// 		_ = s.repo.MarkFailed(ctx, task.ID, err.Error())
// 		return fmt.Errorf("send sms task %d: %w", task.ID, err)
// 	}
// 	return s.repo.MarkSent(ctx, task.ID)
// }

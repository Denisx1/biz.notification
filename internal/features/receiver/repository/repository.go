package receiver_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"notification/internal/core/domain"
	core_errors "notification/internal/core/errors"
	core_postgres_pool "notification/internal/core/repository/postgres/pool"

	"github.com/jackc/pgx/v5"
)

type ReceiverRepository struct {
	pool core_postgres_pool.Pool // или твой core_pgx_pool.Pool, если у него есть QueryRow/Exec
}

func NewReceiverRepository(pool core_postgres_pool.Pool) *ReceiverRepository {
	return &ReceiverRepository{pool: pool}
}

func (r *ReceiverRepository) Get(ctx context.Context, phone string) (domain.SMSVerification, error) {

	const query = `
		SELECT id, code_hash, expires_at
		FROM sms_tasks
		WHERE phone = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var t domain.SMSVerification
	err := r.pool.QueryRow(ctx, query, phone).Scan(&t.ID, &t.CodeHash, &t.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.SMSVerification{}, fmt.Errorf("sms verification not found: %w", core_errors.ErrNotFound)
		}
		return domain.SMSVerification{}, fmt.Errorf("get sms verification: %w", err)
	}
	return t, nil
}

func (r *ReceiverRepository) Create(ctx context.Context, t *domain.SMSTask) (int64, error) {
	var id int64

	query := `
        INSERT INTO sms_tasks (phone, ip, expires_at, code_hash)
        SELECT $1, $2, $3, $4
        WHERE NOT EXISTS (
            SELECT 1 
            FROM sms_tasks 
            WHERE phone = $1 
              AND created_at > NOW() - INTERVAL '60 seconds'
        )
        RETURNING id;
    `

	err := r.pool.QueryRow(
		ctx,
		query,
		t.Phone, t.IP, t.ExpiresAt, t.CodeHash,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Рядок не вставився, бо спрацював ліміт (менше 60 секунд)
			return 0, fmt.Errorf("rate limit exceeded: please wait before sending another sms:%w", core_errors.ErrTooManyRequests)
		}
		return 0, fmt.Errorf("insert sms task: %w", core_errors.ErrTooManyRequests)
	}

	return id, nil
}

// Обновляем только pending: повторная доставка из NATS не перезапишет итоговый статус.
func (r *ReceiverRepository) MarkSent(ctx context.Context, id int64) error {
	return r.setStatus(ctx, id, domain.SMSSent, "")
}

func (r *ReceiverRepository) MarkFailed(ctx context.Context, id int64, reason string) error {
	return r.setStatus(ctx, id, domain.SMSFailed, reason)
}

func (r *ReceiverRepository) MarkExpired(ctx context.Context, id int64) error {
	return r.setStatus(ctx, id, domain.SMSExpired, "")
}

func (r *ReceiverRepository) setStatus(ctx context.Context, id int64, s domain.SMSTaskStatus, reason string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sms_tasks
		SET status = $2,
		    error = $3,
		    sent_at = CASE WHEN $2 = 'sent' THEN now() ELSE sent_at END
		WHERE id = $1 AND status = 'pending'`,
		id, string(s), reason,
	)
	if err != nil {
		return fmt.Errorf("update sms task %d: %w", id, err)
	}
	return nil
}

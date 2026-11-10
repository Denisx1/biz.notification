package receiver_postgres_repository

import (
	"context"
	"fmt"
	receiver_service "notification/internal/features/receiver/service"
)

func (r *ReceiverRepository) RecentCounts(
	ctx context.Context,
	phone, ip string,
) (receiver_service.Counts, error) {
	var c receiver_service.Counts
	query :=
		`
		SELECT
		  count(*) FILTER (WHERE phone = $1 AND created_at > now() - interval '60 seconds'),
		  count(*) FILTER (WHERE phone = $1),
		  count(*) FILTER (WHERE ip = $2 AND $2 <> '')
		FROM sms_tasks
		WHERE created_at > now() - interval '1 hour'
		  AND (phone = $1 OR ip = $2)`
	row := r.pool.QueryRow(
		ctx,
		query,
		phone,
		ip,
	)
	err := row.Scan(&c.PhoneMinute, &c.PhoneHour, &c.IPHour)
	if err != nil {
		return receiver_service.Counts{}, fmt.Errorf("scan error: %w", err)
	}
	return c, err
}

func (r *ReceiverRepository) TakeAttempt(
	ctx context.Context,
	phone string,
) (receiver_service.Attempt, error) {
	var a receiver_service.Attempt
	query := `
		UPDATE sms_tasks SET attempts = attempts + 1
		WHERE id = (
			SELECT id FROM sms_tasks
			WHERE phone = $1
			  AND verified_at IS NULL
			  AND status IN ('pending', 'sent')
			  AND code_hash <> ''
			ORDER BY created_at DESC
			LIMIT 1
		)
		RETURNING id, code_hash, code_expires_at, attempts`
	row := r.pool.QueryRow(ctx, query, phone)
	err := row.Scan(&a.ID, &a.CodeHash, &a.ExpiresAt, &a.Attempts)
	if err != nil {
		return receiver_service.Attempt{}, fmt.Errorf("scan error: %w", err)
	}
	return a, err
}

func (r *ReceiverRepository) MarkVerified(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sms_tasks SET verified_at = now(), status = 'verified', expires_at = NULL WHERE id = $1`, id)
	return err
}
func (r *ReceiverRepository) MarkError(ctx context.Context, id int64, er error) error {
	query := `
		UPDATE sms_tasks SET error = $1, status = 'error' WHERE id = $2
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		er.Error(),
		id,
	)
	return err
}

func (r *ReceiverRepository) LockPhone(
	ctx context.Context,
	phone string,
) error {
	_, err := r.pool.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, phone)
	return err
}

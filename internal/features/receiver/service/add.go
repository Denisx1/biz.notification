package receiver_service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"notification/internal/core/domain"
	receiver_transport_http "notification/internal/features/receiver/transport/http"
	"time"
)

const (
	codeTTL        = 5 * time.Minute
	maxAttempts    = 5
	cooldown       = 60 * time.Second // не чаще 1 SMS в минуту на номер
	phoneHourLimit = 5                // не больше 5 SMS в час на номер
	ipHourLimit    = 20               // не больше 20 SMS в час с одного IP
)

type Counts struct {
	PhoneMinute,
	PhoneHour,
	IPHour int
}
type Attempt struct {
	ID        int64
	CodeHash  string
	ExpiresAt time.Time
	Attempts  int
}

func (s *ReceiverService) Add(
	ctx context.Context,
	req receiver_transport_http.AddTaskRequest,
) error {

	code, err := generateCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	taskDomain := domain.NewSMSUninitialized(
		req.Phone,
		req.Expiration,
		s.hashCode(code),
	)

	id, err := s.repo.Create(ctx, taskDomain)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(domain.SMSMessage{
		ID:    id,
		Code:  code,
		Phone: taskDomain.Phone,
	})

	if err := s.publisher.Publish(ctx, "sms.send", payload); err != nil {
		_ = s.repo.MarkFailed(ctx, id, "publish: "+err.Error())
		return fmt.Errorf("publish task %d: %w", taskDomain.ID, err)
	}
	if err := s.repo.MarkSent(ctx, id); err != nil {
		return fmt.Errorf("publish task %d: %w", taskDomain.ID, err)
	}
	return nil
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}
func (s *ReceiverService) hashCode(code string) string {
	m := hmac.New(sha256.New, []byte("secret"))
	m.Write([]byte(code))
	return hex.EncodeToString(m.Sum(nil))
}

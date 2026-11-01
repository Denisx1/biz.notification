package receiver_service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	core_errors "notification/internal/core/errors"
	"time"
)

func (s *ReceiverService) Verify(
	ctx context.Context,
	phone, code string,
) error {
	task, err := s.repo.Get(ctx, phone)
	if err != nil {
		return err
	}
	if time.Now().After(task.ExpiresAt) {
		return core_errors.ErrCodeExpired
	}
	if !s.verifyCode(code, task.CodeHash) {
		return core_errors.ErrInvalidCode
	}
	if err := s.repo.MarkVerified(ctx, task.ID); err != nil {
		return err
	}
	return nil
}

func (s *ReceiverService) verifyCode(code, codehashed string) bool {
	stored, err := hex.DecodeString(codehashed)
	if err != nil {
		return false
	}
	m := hmac.New(sha256.New, []byte("secret"))
	m.Write([]byte(code))
	return hmac.Equal(m.Sum(nil), stored)

}

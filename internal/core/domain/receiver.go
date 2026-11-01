package domain

import "time"

type SMSTaskStatus string

const (
	SMSPending SMSTaskStatus = "pending"
	SMSSent    SMSTaskStatus = "sent"
	SMSFailed  SMSTaskStatus = "failed"
	SMSExpired SMSTaskStatus = "expired"
)

type SMSTask struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	IP        string    `json:"ip"`
	IsUsed    bool      `json:"is_used"`
	ExpiresAt time.Time `json:"expiresAt"` // от клиента: когда задача протухает
	CodeHash  string    `json:"-"`         // в NATS не уходит
}

type SMSMessage struct {
	ID    int64
	Code  string
	Phone string
}

func NewSmsTask(
	id int64,
	phone string,
	ip string,
	expiresAt time.Time,
	codeHash string,
) *SMSTask {
	return &SMSTask{
		ID:        id,
		Phone:     phone,
		IP:        ip,
		ExpiresAt: expiresAt,
		CodeHash:  codeHash,
	}
}
func NewSMSUninitialized(
	phone string,
	ip string,
	expiresAt int64,
	codeHash string,

) *SMSTask {
	expireAt := time.Unix(expiresAt, 0)
	return NewSmsTask(
		int64(UninitializedID),
		phone,
		ip,
		expireAt,
		codeHash,
	)
}

func (t SMSTask) Expired() bool {
	return !t.ExpiresAt.IsZero() && time.Now().After(t.ExpiresAt)
}

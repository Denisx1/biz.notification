package sender_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"notification/internal/core/domain"
)

const baseURL = "https://api.sendpulse.com"

type SMSRequest struct {
	Sender               string            `json:"sender"`
	Phones               []string          `json:"phones"`
	Body                 string            `json:"body"`
	Route                map[string]string `json:"route,omitempty"`
	Emulate              bool              `json:"emulate"`
	Date                 string            `json:"date,omitempty"`
	StatLinkTracking     bool              `json:"stat_link_tracking"`
	StatLinkNeedProtocol bool              `json:"stat_link_need_protocol"`
}

func (s *SenderService) SendSMS(sms domain.SMSMessage) (string, error) {

	smsPayload := SMSRequest{
		Sender:               "CherkasyBiz",
		Phones:               []string{sms.Phone},
		Body:                 fmt.Sprintf("%s %s", sms.Code, s.baseSMS),
		StatLinkTracking:     false,
		StatLinkNeedProtocol: false,
	}

	payload, err := json.Marshal(smsPayload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/sms/send", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	out, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(out))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sendpulse error %d: %s", resp.StatusCode, out)
	}
	return string(out), nil
}

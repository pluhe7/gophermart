package services

import (
	"fmt"
	"io"
	"net/http"
)

type AccrualSystemOrderStatus string

const (
	AccrualSystemOrderStatusRegistered = "REGISTERED"
	AccrualSystemOrderStatusProcessing = "PROCESSING"
	AccrualSystemOrderStatusInvalid    = "INVALID"
	AccrualSystemOrderStatusProcessed  = "PROCESSED"
)

var AccrualSystemQueryLimit = 100

type AccrualSystemResponse struct {
	Order   string                   `json:"order"`
	Status  AccrualSystemOrderStatus `json:"status"`
	Accrual *float64                 `json:"accrual,omitempty"`
}

type AccrualSystem struct {
	serviceURL string
}

func NewAccrualSystem(url string) *AccrualSystem {
	return &AccrualSystem{
		serviceURL: url,
	}
}

func (as *AccrualSystem) GetOrder(number string) ([]byte, int, error) {
	reqURL := fmt.Sprintf("%s/api/orders/%s", as.serviceURL, number)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("new request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do request: %w", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read body: %w", err)
	}

	return body, res.StatusCode, nil
}

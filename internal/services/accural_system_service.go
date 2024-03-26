package services

import (
	"encoding/json"
	"errors"
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

const AccrualSystemQueryLimit = 100

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

func (as *AccrualSystem) GetOrder(number string) (*AccrualSystemResponse, error) {
	reqURL := fmt.Sprintf("%s/api/orders/%s", as.serviceURL, number)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	switch res.StatusCode {
	case http.StatusOK:
		var response AccrualSystemResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, fmt.Errorf("unmarshal response body: %w", err)
		}

		return &response, nil

	case http.StatusNoContent:
		return nil, errors.New("order not registered")

	case http.StatusTooManyRequests:
		return nil, errors.New(string(body))

	case http.StatusInternalServerError:
		return nil, fmt.Errorf("service error: %s", string(body))

	default:
		return nil, fmt.Errorf("unexpected response status %d %s", res.StatusCode, string(body))
	}
}

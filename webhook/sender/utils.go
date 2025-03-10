package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"webhook/logging"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

var HTTPClient HTTPDoer = &http.Client{}

var marshalJSON = func(data interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error marshaling JSON: %s", err))
		return nil, err
	}
	return jsonBytes, nil
}

var prepareRequest = func(url string, jsonBytes []byte, secretHash string) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error during the webhook request preparation"))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if secretHash != "" {
		req.Header.Set("X-Secret-Hash", secretHash)
	}

	return req, nil
}

var sendRequest = func(req *http.Request) (*http.Response, error) {
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

var closeResponse = func(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error closing response body: %s", err))
	}
}

var processResponse = func(resp *http.Response) (string, []byte, error) {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.WebhookLogger(logging.ErrorType, fmt.Errorf("error reading response body: %s", err))
		return "failed", nil, err
	}

	status := "failed"
	if resp.StatusCode == http.StatusOK {
		status = "delivered"
	}

	if status == "failed" {
		logging.WebhookLogger(logging.WarningType, fmt.Errorf("HTTP request failed with status code: %d, response body: %s", resp.StatusCode, string(respBody)))
	}

	return status, respBody, nil
}

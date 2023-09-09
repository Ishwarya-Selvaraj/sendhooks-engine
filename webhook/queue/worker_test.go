package queue

import (
	"context"
	"testing"
	"time"
	redisClient "webhook/redis"
)

var (
	// Mocks
	sendWebhookCalls            int
	webhookLoggerCalls          int
	sendWebhook                 = func(data interface{}, url string, webhookId string, secretHash string) error { return nil }
	sendWebhookWithRetriesCalls int
)

func TestProcessWebhooks(t *testing.T) {
	sendWebhookWithRetriesCalls = 0
	webhookQueue := make(chan redisClient.WebhookPayload, 5)

	// Mock
	sendWebhookWithRetries = func(payload redisClient.WebhookPayload) {
		sendWebhookWithRetriesCalls++
	}

	go ProcessWebhooks(context.TODO(), webhookQueue)

	webhookQueue <- redisClient.WebhookPayload{WebhookId: "1"}
	webhookQueue <- redisClient.WebhookPayload{WebhookId: "2"}
	close(webhookQueue)

	if sendWebhookWithRetriesCalls != 2 {
		t.Fatalf("Expected sendWebhookWithRetries to be called 2 times, but got %d calls", sendWebhookWithRetriesCalls)
	}
}

func TestSendWebhookWithRetries(t *testing.T) {
	webhookLoggerCalls = 0
	payload := redisClient.WebhookPayload{WebhookId: "1"}

	// Mock
	retryWithExponentialBackoff = func(payload redisClient.WebhookPayload) error {
		return nil
	}

	sendWebhookWithRetries(payload)

	if webhookLoggerCalls != 0 {
		t.Fatalf("Expected no logging, but got %d logs", webhookLoggerCalls)
	}
}

func TestCalculateBackoff(t *testing.T) {
	if backoff := calculateBackoff(initialBackoff); backoff != 2*time.Second {
		t.Fatalf("Expected backoff to be 2 seconds, but got %v", backoff)
	}

	if backoff := calculateBackoff(maxBackoff / 2); backoff != maxBackoff {
		t.Fatalf("Expected backoff to be %v, but got %v", maxBackoff, backoff)
	}
}

func TestRetryWithExponentialBackoff(t *testing.T) {
	sendWebhookCalls = 0
	webhookLoggerCalls = 0

	payload := redisClient.WebhookPayload{WebhookId: "1"}

	// Mock
	sendWebhook = func(data interface{}, url string, webhookId string, secretHash string) error {
		sendWebhookCalls++
		return nil
	}

	retryWithExponentialBackoff(payload)

	if sendWebhookCalls != maxRetries {
		t.Fatalf("Expected SendWebhook to be called %d times, but got %d calls", maxRetries, sendWebhookCalls)
	}

	if webhookLoggerCalls != maxRetries {
		t.Fatalf("Expected WebhookLogger to be called %d times, but got %d calls", maxRetries, webhookLoggerCalls)
	}
}

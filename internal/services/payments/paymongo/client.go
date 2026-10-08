package paymongo

import (
	"bytes"
	"context"
	"fmt"
	"gobrewflow/internal/config"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL     string
	secretKey   string
	http        *http.Client
	retryConfig config.RetryConfig
	log         *slog.Logger
}

func NewClient(
	baseURL string,
	secretKey string,
	httpClient *http.Client,
	retryCfg config.RetryConfig,
	log *slog.Logger,
) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		secretKey:   secretKey,
		http:        httpClient,
		retryConfig: retryCfg,
		log:         log,
	}
}

func (c *Client) DoCreateCheckout(
	ctx context.Context,
	body []byte,
	idempotencyKey string,
) ([]byte, error) {
	maxAttempts := c.retryConfig.MaxAttempts

	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {

		// if attempt is greater than 1, we wait for a backoff duration before retrying
		if attempt > 1 {
			// wait for a backoff duration before retrying
			// it uses channels and timers to wait for the backoff duration or until the context is canceled
			if err := c.waitForRetry(ctx, attempt); err != nil {
				return nil, err
			}
		}

		// else we create a new request and send it to the PayMongo API
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.baseURL+"/v2/checkout_sessions",
			bytes.NewReader(body),
		)
		if err != nil {
			return nil, err
		}

		// set the required headers for the request, including the idempotency key
		req.SetBasicAuth(c.secretKey, "")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		resp, err := c.http.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			// if the request failed and we have more attempts left, we log a warning and continue to the next attempt
			if attempt < maxAttempts {
				c.log.WarnContext(
					ctx,
					"paymongo request retrying",
					slog.Int("attempt", attempt),
					slog.Int("max_attempts", maxAttempts),
					slog.String("error", err.Error()),
				)

				continue // retry
			}

			// return an error if the request failed and we have no more attempts left
			return nil, fmt.Errorf(
				"paymongo: request failed after %d attempts: %w",
				attempt,
				err,
			)
		}

		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			return nil, readErr
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return responseBody, nil
		}

		if isRetryableStatus(resp.StatusCode) && attempt < maxAttempts {
			c.log.WarnContext(
				ctx,
				"paymongo request retrying",
				slog.Int("attempt", attempt),
				slog.Int("max_attempts", maxAttempts),
				slog.Int("status", resp.StatusCode),
			)
			continue
		}
		return nil, fmt.Errorf(
			"paymongo: request failed after %d attempts: status %d: %s",
			attempt,
			resp.StatusCode,
			string(responseBody),
		)
	}

	return nil, fmt.Errorf("paymongo: request failed")
}

func (c *Client) waitForRetry(
	ctx context.Context,
	attempt int,
) error {
	delay := c.calculateBackoff(attempt)

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	// wait for either the context to be canceled or the timer to expire
	case <-ctx.Done():
		return ctx.Err()
	// wait for the timer to expire before retrying
	case <-timer.C:
		return nil
	}
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	// Prevent panic if BaseDelay or configuration wasn't provided or set properly
	if c.retryConfig.BaseDelay <= 0 {
		return 500 * time.Millisecond
	}

	// Exponential backoff formula: base * 2^(attempt-1)
	// e.g (1st attempt: base delay (500ms), 2nd attempt: 1000ms, 3rd attempt: 2000ms, etc.)
	backoff := c.retryConfig.BaseDelay * (1 << uint(attempt-1))

	if backoff > c.retryConfig.MaxDelay || backoff <= 0 {
		backoff = c.retryConfig.MaxDelay
	}

	// Add Jitter (randomization between 0 and the current backoff)
	// to prevent a "thundering herd" effect on the server.
	// e.g (if backoff is 1000ms, jitter will be a random value between 0 and 1000ms)
	jitter := rand.Int63n(int64(backoff))
	return time.Duration(jitter)
}

func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case
		http.StatusRequestTimeout,      // 408
		http.StatusTooManyRequests,     // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

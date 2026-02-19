package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

var client = http.Client{
	Timeout: time.Second * 20,
}

func SendRequest(ctx context.Context, method, url string, headers map[string]string, body []byte, query map[string]string, respModel any) error {
	logger := zerolog.Ctx(ctx)

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if user, ok1 := headers["Username"]; ok1 {
		if pass, ok2 := headers["Password"]; ok2 {
			logger.Warn().Msg("Setting username and password")
			req.SetBasicAuth(user, pass)
		}
		delete(headers, "Username")
		delete(headers, "Password")
	}

	// Apply all remaining headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if len(body) > 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		logger.Error().Err(err).Str("method", method).Str("url", url).Msg("HTTP response error")
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		data, er := io.ReadAll(resp.Body)
		if er != nil {
			logger.Error().Err(er).Int("status", resp.StatusCode).Msg("HTTP response error")
			return fmt.Errorf("failed to read response body: %w", er)
		}

		err := fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
		logger.Error().Err(err).Str("method", method).Str("url", url).Int("status", resp.StatusCode).Msg("HTTP response error")
		return err
	}

	if respModel != nil {
		err = json.NewDecoder(resp.Body).Decode(respModel)
		if err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func FirstNonEmpty(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

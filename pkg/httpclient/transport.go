package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dev-au/CodeStream/pkg/customerrors"
	"github.com/dev-au/CodeStream/pkg/logs"
)

func NewTransport() (*Transport, func()) {
	logs.Info("Creating new HTTP transport")

	transport := &Transport{
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}

	cleanup := func() {
		logs.Info("Cleaning up HTTP transport")
		transport.HTTPClient.CloseIdleConnections()
	}

	return transport, cleanup
}

func (t *Transport) doRequest(ctx context.Context, url, method string, body []byte, queryString *JsonMap, headers []*Header) ([]byte, int, error) {
	if t.BeforeRequest != nil {
		url, body, queryString, headers = t.BeforeRequest(ctx, url, body, queryString, headers)
	}
	logs.InfoCtx(ctx, "Preparing HTTP request",
		"method", method,
		"url", url,
		"bodySize", len(body),
		"hasQueryString", queryString != nil && len(*queryString) > 0,
	)

	logs.DebugCtx(ctx, "Request details",
		"method", method,
		"url", url,
		"body", string(body),
		"queryString", queryString,
		"extraHeaders", headers,
	)

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		logs.PanicCtx(ctx, "Failed to create HTTP request", "error", err, "url", url, "method", method)
		return nil, 0, customerrors.WrapSystemError(err)
	}

	for _, header := range headers {
		req.Header.Set(header.Key, header.Value)
		logs.DebugCtx(ctx, "Setting extra header", "key", header.Key, "value", header.Value)
	}

	if queryString != nil && len(*queryString) > 0 {
		query := req.URL.Query()
		for key, value := range *queryString {
			query.Add(key, fmt.Sprintf("%v", value))
			logs.DebugCtx(ctx, "Adding query parameter", "key", key, "value", value)
		}
		req.URL.RawQuery = query.Encode()
		logs.DebugCtx(ctx, "Final query string", "rawQuery", req.URL.RawQuery)
	}

	startTime := time.Now()
	logs.InfoCtx(ctx, "Sending HTTP request", "method", method, "url", url)

	resp, err := t.HTTPClient.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		logs.ErrorCtx(ctx, "HTTP request failed",
			"error", err,
			"url", url,
			"method", method,
			"duration", duration,
		)
		return nil, 0, customerrors.WrapSystemError(err)
	}
	defer func() { _ = resp.Body.Close() }()

	logs.InfoCtx(ctx, "Received HTTP response",
		"statusCode", resp.StatusCode,
		"status", resp.Status,
		"duration", duration,
		"url", url,
	)

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorCtx(ctx, "Failed to read response body",
			"error", err,
			"statusCode", resp.StatusCode,
			"url", url,
		)
		return nil, resp.StatusCode, customerrors.WrapSystemError(err)
	}

	logs.DebugCtx(ctx, "Response body received",
		"statusCode", resp.StatusCode,
		"contentSize", len(content),
		"content", string(content),
	)
	switch {
	case resp.StatusCode > http.StatusInternalServerError:
		logs.ErrorCtx(ctx, "HTTP request returned server error status",
			"statusCode", resp.StatusCode,
			"status", resp.Status,
			"url", url,
			"responseBody", string(content),
		)
		err = customerrors.WrapSystemError(errors.New(string(content)))
	case resp.StatusCode >= http.StatusBadRequest:
		logs.WarnCtx(ctx, "HTTP request returned client error status",
			"statusCode", resp.StatusCode,
			"status", resp.Status,
			"url", url,
			"responseBody", string(content),
		)
		err = customerrors.WrapExternalServiceError(errors.New(string(content)))
	default:
		logs.InfoCtx(ctx, "HTTP request completed successfully",
			"statusCode", resp.StatusCode,
			"duration", duration,
			"responseSize", len(content),
			"url", url,
		)
	}

	return content, resp.StatusCode, err
}

func (t *Transport) Get(ctx context.Context, url string, queryString *JsonMap, headers []*Header) ([]byte, int, error) {
	logs.InfoCtx(ctx, "Executing GET request", "url", url, "hasQueryString", queryString != nil && len(*queryString) > 0)
	return t.doRequest(ctx, url, http.MethodGet, nil, queryString, headers)
}

func (t *Transport) Post(ctx context.Context, url string, body any, headers []*Header) ([]byte, int, error) {
	logs.InfoCtx(ctx, "Executing POST request", "url", url, "bodyType", fmt.Sprintf("%T", body))

	var requestBody []byte
	var err error

	switch v := body.(type) {
	case *JsonMap:
		requestBody, err = json.Marshal(*v)
		headers = append(headers, &Header{"Content-Type", "application/json"})
	case JsonMap:
		requestBody, err = json.Marshal(v)
		headers = append(headers, &Header{"Content-Type", "application/json"})
	case map[string]any:
		requestBody, err = json.Marshal(v)
		headers = append(headers, &Header{"Content-Type", "application/json"})
	case []byte:
		requestBody = v
	default:
		err = fmt.Errorf("unsupported body type %T", body)
	}

	if err != nil {
		logs.PanicCtx(ctx, "Failed to prepare POST request body", "error", err, "url", url)
		return nil, 0, customerrors.WrapSystemError(err)
	}
	return t.doRequest(ctx, url, http.MethodPost, requestBody, nil, headers)
}

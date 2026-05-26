package httpclient

import (
	"context"
	"net/http"
)

type JsonMap map[string]any

type Header struct {
	Key   string
	Value string
}
type Transport struct {
	HTTPClient    *http.Client
	BeforeRequest func(ctx context.Context, url string, body []byte, queryString *JsonMap, headers []*Header) (string, []byte, *JsonMap, []*Header)
}

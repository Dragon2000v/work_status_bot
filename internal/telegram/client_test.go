package telegram

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestClientSendMessageSuccessAndFailure(t *testing.T) {
	calls := 0
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewBufferString("{}"))}, nil
		}
		return &http.Response{StatusCode: http.StatusBadGateway, Status: "502 Bad Gateway", Body: io.NopCloser(bytes.NewBufferString("{}"))}, nil
	})}
	client := NewClient("token", httpClient)
	client.baseURL = "https://telegram.test"
	if err := client.SendMessage(context.Background(), 1, "ok"); err != nil {
		t.Fatalf("send ok: %v", err)
	}
	if err := client.SendMessage(context.Background(), 1, "fail"); err == nil {
		t.Fatalf("want failure")
	}
}

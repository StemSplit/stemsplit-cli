package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := New("sk_test")
	c.baseURL = srv.URL
	return c
}

func TestUploadFileSendsSignedContentType(t *testing.T) {
	var gotType, gotBody, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		gotType = r.Header.Get("Content-Type")
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "song.mp3")
	if err := os.WriteFile(path, []byte("audio"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New("sk_test").UploadFile(srv.URL, path, "audio/mpeg"); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if gotType != "audio/mpeg" {
		t.Errorf("Content-Type = %q, want audio/mpeg", gotType)
	}
	if gotAuth != "" {
		t.Errorf("presigned upload must not send Authorization, got %q", gotAuth)
	}
	if gotBody != "audio" {
		t.Errorf("body = %q, want file contents", gotBody)
	}
}

func TestUploadFileReportsStorageError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("SignatureDoesNotMatch"))
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "song.mp3")
	_ = os.WriteFile(path, []byte("audio"), 0o600)

	err := New("sk_test").UploadFile(srv.URL, path, "audio/mpeg")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("err = %v, want 403 upload error", err)
	}
}

func TestGetUploadURLParsesContentType(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/upload" || r.Method != http.MethodPost {
			t.Errorf("got %s %s, want POST /upload", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"uploadUrl":"https://r2/x","uploadKey":"uploads/1/input.mp3","expiresAt":"t","contentType":"audio/mpeg"}`))
	})

	resp, err := c.GetUploadURL("song.mp3", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ContentType != "audio/mpeg" || resp.UploadKey != "uploads/1/input.mp3" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestAPIRequestHeaders(t *testing.T) {
	SetVersion("1.2.3")
	t.Cleanup(func() { version = "dev" })

	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer sk_test" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "stemsplit-cli/1.2.3" {
			t.Errorf("User-Agent = %q", got)
		}
		_, _ = w.Write([]byte(`{"balanceSeconds":125,"balanceMinutes":2,"balanceFormatted":"2:05","updatedAt":"t"}`))
	})

	resp, err := c.GetBalance()
	if err != nil {
		t.Fatal(err)
	}
	if resp.BalanceSeconds != 125 {
		t.Errorf("BalanceSeconds = %d, want 125", resp.BalanceSeconds)
	}
}

func TestAPIErrorMessageIsSurfaced(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":"INVALID_API_KEY","message":"Invalid API key."}}`))
	})

	_, err := c.GetBalance()
	if err == nil || err.Error() != "Invalid API key." {
		t.Fatalf("err = %v, want API message", err)
	}
}

func TestNonJSONErrorFallsBackToStatus(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	})

	_, err := c.GetJob("abc")
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("err = %v, want status in message", err)
	}
}

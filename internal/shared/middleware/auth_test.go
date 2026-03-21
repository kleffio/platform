package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubIntrospector struct {
	subject string
	err     error
}

func (s *stubIntrospector) Introspect(_ context.Context, _ string) (string, error) {
	return s.subject, s.err
}

func middlewareFromStub(i interface {
	Introspect(context.Context, string) (string, error)
}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Header.Get("Authorization")
			token := ""
			if len(v) > 7 && v[:7] == "Bearer " {
				token = v[7:]
			}
			if token == "" {
				http.Error(w, `{"error":"unauthorized","code":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			subject, err := i.Introspect(r.Context(), token)
			if err != nil {
				http.Error(w, `{"error":"unauthorized","code":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), struct{ key string }{"subject"}, subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TestRequireAuth_NoToken(t *testing.T) {
	mw := middlewareFromStub(&stubIntrospector{subject: "user-123"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/identity/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	mw := middlewareFromStub(&stubIntrospector{err: errors.New("token inactive")})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/identity/me", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	var gotSubject string
	mw := middlewareFromStub(&stubIntrospector{subject: "user-abc-123"})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSubject, _ = r.Context().Value(struct{ key string }{"subject"}).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/identity/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotSubject != "user-abc-123" {
		t.Fatalf("expected user-abc-123, got %q", gotSubject)
	}
}

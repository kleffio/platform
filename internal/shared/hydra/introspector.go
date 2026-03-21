package hydra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Introspector struct {
	adminURL string
	client   *http.Client
}

func NewIntrospector(adminURL string) *Introspector {
	return &Introspector{
		adminURL: strings.TrimRight(adminURL, "/"),
		client:   &http.Client{},
	}
}

type introspectResponse struct {
	Active  bool   `json:"active"`
	Subject string `json:"sub"`
	Email   string `json:"email"`
}

func (i *Introspector) Introspect(ctx context.Context, token string) (string, error) {
	endpoint := i.adminURL + "/admin/oauth2/introspect"

	body := url.Values{}
	body.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("build introspect request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := i.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("introspect request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("introspect returned %d", resp.StatusCode)
	}

	var result introspectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode introspect response: %w", err)
	}

	if !result.Active {
		return "", fmt.Errorf("token inactive or expired")
	}

	if result.Subject == "" {
		return "", fmt.Errorf("token missing subject claim")
	}

	return result.Subject, nil
}

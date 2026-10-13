// Package mattermost is a minimal client for the parts of the Mattermost REST API v4
// and WebSocket API that the bridge uses.
package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	ServerURL string
	Token     string
	HTTP      *http.Client
	UserAgent string

	SessionCheckInterval time.Duration
}

type APIError struct {
	StatusCode int    `json:"status_code"`
	ID         string `json:"id"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("HTTP %d: %s (%s)", e.StatusCode, e.Message, e.ID)
}

// NormalizeServerURL turns user input like "chat.example.org/" into "https://chat.example.org".
func NormalizeServerURL(input string) (string, error) {
	input = strings.TrimSpace(input)
	if !strings.Contains(input, "://") {
		input = "https://" + input
	}
	u, err := url.Parse(input)
	if err != nil {
		return "", err
	} else if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("unsupported URL scheme %q", u.Scheme)
	} else if u.Host == "" {
		return "", fmt.Errorf("missing host in server URL")
	}
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimSuffix(u.String(), "/"), nil
}

// Login creates a session with a username or email and a password and stores its token in c.
func (c *Client) Login(ctx context.Context, loginID, password string) (*User, error) {
	body := map[string]string{"login_id": loginID, "password": password}
	resp, err := c.do(ctx, http.MethodPost, "/users/login", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	c.Token = resp.Header.Get("Token")
	if c.Token == "" {
		return nil, fmt.Errorf("login response didn't include a session token")
	}
	var user User
	return &user, json.NewDecoder(resp.Body).Decode(&user)
}

func (c *Client) Logout(ctx context.Context) error {
	resp, err := c.do(ctx, http.MethodPost, "/users/logout", nil)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (c *Client) GetMe(ctx context.Context) (*User, error) {
	var user User
	return &user, c.getJSON(ctx, "/users/me", &user)
}

func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	var user User
	return &user, c.getJSON(ctx, "/users/"+url.PathEscape(userID), &user)
}

func (c *Client) GetChannel(ctx context.Context, channelID string) (*Channel, error) {
	var channel Channel
	return &channel, c.getJSON(ctx, "/channels/"+url.PathEscape(channelID), &channel)
}

func (c *Client) GetProfileImage(ctx context.Context, userID string) ([]byte, error) {
	resp, err := c.do(ctx, http.MethodGet, "/users/"+url.PathEscape(userID)+"/image", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(io.LimitReader(resp.Body, 16<<20))
}

func (c *Client) getJSON(ctx context.Context, path string, into any) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(into)
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.ServerURL+"/api/v4"+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.setHeaders(req.Header)
	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		apiErr := &APIError{}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(apiErr)
		apiErr.StatusCode = resp.StatusCode
		return nil, apiErr
	}
	return resp, nil
}

func (c *Client) setHeaders(h http.Header) {
	if c.Token != "" {
		h.Set("Authorization", "Bearer "+c.Token)
	}
	if c.UserAgent != "" {
		h.Set("User-Agent", c.UserAgent)
	}
}

package mattermost

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeServerURL(t *testing.T) {
	for input, expected := range map[string]string{
		"chat.example.org":              "https://chat.example.org",
		" https://chat.example.org/ ":   "https://chat.example.org",
		"http://localhost:8065":         "http://localhost:8065",
		"https://example.org/mm/?x=1#y": "https://example.org/mm",
	} {
		actual, err := NormalizeServerURL(input)
		require.NoError(t, err, input)
		assert.Equal(t, expected, actual, input)
	}
	for _, input := range []string{"ftp://example.org", "https://"} {
		_, err := NormalizeServerURL(input)
		assert.Error(t, err, input)
	}
}

func TestLogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		if r.URL.Path != "/api/v4/users/login" || body["password"] != "correct" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"id":"api.user.login.invalid_credentials_email_username","message":"Enter a valid email or username and/or password."}`))
			return
		}
		w.Header().Set("Token", "session-token")
		_, _ = w.Write([]byte(`{"id":"u1","username":"robin","first_name":"Robin"}`))
	}))
	defer srv.Close()

	cli := &Client{ServerURL: srv.URL}
	_, err := cli.Login(context.Background(), "robin", "wrong")
	require.Error(t, err)
	assert.True(t, IsUnauthorized(err))
	assert.Empty(t, cli.Token)

	user, err := cli.Login(context.Background(), "robin", "correct")
	require.NoError(t, err)
	assert.Equal(t, "session-token", cli.Token)
	assert.Equal(t, "u1", user.ID)
	assert.Equal(t, "Robin", user.DisplayName())
}

func TestListen(t *testing.T) {
	post := Post{ID: "p1", ChannelID: "c1", UserID: "u2", Message: "hello", CreateAt: 1760000000000}
	postJSON, _ := json.Marshal(post)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer good" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		conn, err := websocket.Accept(w, r, nil)
		require.NoError(t, err)
		defer conn.CloseNow()
		ctx := r.Context()
		hello, _ := json.Marshal(map[string]any{"event": "hello", "data": map[string]string{}, "seq": 0})
		data, _ := json.Marshal(PostedData{ChannelType: ChannelOpen, Post: string(postJSON)})
		posted, _ := json.Marshal(WebSocketEvent{Event: EventPosted, Data: data, Seq: 1})
		_ = conn.Write(ctx, websocket.MessageText, hello)
		_ = conn.Write(ctx, websocket.MessageText, posted)
		_, _, _ = conn.Read(ctx)
	}))
	defer srv.Close()

	err := (&Client{ServerURL: srv.URL, Token: "bad"}).Listen(context.Background(), func() {}, func(*WebSocketEvent) {})
	assert.True(t, IsUnauthorized(err), "expected 401, got %v", err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connected := false
	var received *Post
	err = (&Client{ServerURL: srv.URL, Token: "good"}).Listen(ctx, func() { connected = true }, func(evt *WebSocketEvent) {
		var data PostedData
		require.NoError(t, json.Unmarshal(evt.Data, &data))
		received, err = data.ParsePost()
		require.NoError(t, err)
		cancel()
	})
	assert.ErrorIs(t, err, context.Canceled)
	assert.True(t, connected)
	assert.Equal(t, &post, received)
}

func TestListenDetectsRevokedSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/users/me" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		conn, err := websocket.Accept(w, r, nil)
		require.NoError(t, err)
		defer conn.CloseNow()
		// Like Mattermost, keep the socket open without sending anything.
		_, _, _ = conn.Read(r.Context())
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := (&Client{ServerURL: srv.URL, Token: "revoked", SessionCheckInterval: 50 * time.Millisecond}).Listen(ctx, func() {}, func(*WebSocketEvent) {})
	assert.True(t, IsUnauthorized(err), "expected 401, got %v", err)
}

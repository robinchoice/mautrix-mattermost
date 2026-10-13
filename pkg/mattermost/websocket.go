package mattermost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
)

const pingInterval = 30 * time.Second

// defaultSessionCheckInterval is how often Listen verifies the session token over REST.
// Mattermost keeps a websocket open after its session is revoked and just stops sending events.
const defaultSessionCheckInterval = 5 * time.Minute

// Listen connects to the WebSocket API and calls handler for every event until ctx is canceled
// or the connection fails. onConnected is called once the server has said hello.
// An *APIError with StatusCode 401 means the session token is no longer valid.
func (c *Client) Listen(ctx context.Context, onConnected func(), handler func(*WebSocketEvent)) error {
	wsURL := "ws" + strings.TrimPrefix(c.ServerURL, "http") + "/api/v4/websocket"
	header := http.Header{}
	c.setHeaders(header)
	conn, resp, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: header, HTTPClient: c.HTTP})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return &APIError{StatusCode: resp.StatusCode, Message: "session token rejected"}
		}
		return fmt.Errorf("failed to connect websocket: %w", err)
	}
	defer conn.CloseNow()
	conn.SetReadLimit(4 << 20)

	ctx, cancel := context.WithCancel(ctx)
	keepaliveErr := make(chan error, 1)
	keepaliveDone := make(chan struct{})
	defer func() {
		cancel()
		<-keepaliveDone
	}()
	go func() {
		defer close(keepaliveDone)
		err := c.keepalive(ctx, conn)
		if err != nil && ctx.Err() == nil {
			keepaliveErr <- err
			conn.CloseNow()
		}
	}()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			select {
			case kerr := <-keepaliveErr:
				return kerr
			default:
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Mattermost accepts the websocket with a revoked token and then just closes it.
			if _, meErr := c.GetMe(ctx); IsUnauthorized(meErr) {
				return meErr
			}
			return err
		}
		var evt WebSocketEvent
		if err = json.Unmarshal(data, &evt); err != nil {
			return fmt.Errorf("failed to parse websocket message: %w", err)
		}
		switch evt.Event {
		case "":
			// Response to a client action, we don't send any.
		case "hello":
			onConnected()
		default:
			handler(&evt)
		}
	}
}

func (c *Client) keepalive(ctx context.Context, conn *websocket.Conn) error {
	pingTicker := time.NewTicker(pingInterval)
	defer pingTicker.Stop()
	interval := c.SessionCheckInterval
	if interval == 0 {
		interval = defaultSessionCheckInterval
	}
	sessionTicker := time.NewTicker(interval)
	defer sessionTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-pingTicker.C:
			pingCtx, cancel := context.WithTimeout(ctx, pingInterval)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return fmt.Errorf("ping failed: %w", err)
			}
		case <-sessionTicker.C:
			if _, err := c.GetMe(ctx); IsUnauthorized(err) {
				return err
			}
		}
	}
}

func IsUnauthorized(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized
}

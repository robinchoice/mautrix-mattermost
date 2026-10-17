package connector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maunium.net/go/mautrix/bridgev2"
)

func TestCredentialsLoginServerStep(t *testing.T) {
	mm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/system/ping" {
			_, _ = w.Write([]byte(`{"status":"OK"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mm.Close()
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html></html>`))
	}))
	defer other.Close()

	login := &CredentialsLogin{}
	step, err := login.Start(context.Background())
	require.NoError(t, err)
	require.Len(t, step.UserInputParams.Fields, 1)
	assert.Equal(t, "server_url", step.UserInputParams.Fields[0].ID)

	_, err = login.SubmitUserInput(context.Background(), map[string]string{"server_url": other.URL})
	var respErr bridgev2.RespError
	require.ErrorAs(t, err, &respErr)
	assert.Nil(t, login.client)

	step, err = login.SubmitUserInput(context.Background(), map[string]string{"server_url": mm.URL + "/"})
	require.NoError(t, err)
	require.Len(t, step.UserInputParams.Fields, 2)
	assert.Equal(t, "username", step.UserInputParams.Fields[0].ID)
	assert.Equal(t, bridgev2.LoginInputFieldTypePassword, step.UserInputParams.Fields[1].Type)
	assert.Equal(t, mm.URL, login.client.ServerURL)
}

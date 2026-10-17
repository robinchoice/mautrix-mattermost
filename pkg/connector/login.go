package connector

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/status"

	"github.com/robinchoice/mautrix-mattermost/pkg/mattermost"
)

const LoginFlowCredentials = "credentials"

var loginFlows = []bridgev2.LoginFlow{{
	Name:        "Username and password",
	Description: "Log in with your Mattermost server URL, then your username or email and password.",
	ID:          LoginFlowCredentials,
}}

func (mc *MattermostConnector) GetLoginFlows() []bridgev2.LoginFlow {
	return loginFlows
}

func (mc *MattermostConnector) CreateLogin(ctx context.Context, user *bridgev2.User, flowID string) (bridgev2.LoginProcess, error) {
	if flowID != LoginFlowCredentials {
		return nil, fmt.Errorf("unknown login flow %q", flowID)
	}
	return &CredentialsLogin{User: user, Main: mc}, nil
}

type CredentialsLogin struct {
	User *bridgev2.User
	Main *MattermostConnector

	client *mattermost.Client
}

var _ bridgev2.LoginProcessUserInput = (*CredentialsLogin)(nil)

func (cl *CredentialsLogin) Start(ctx context.Context) (*bridgev2.LoginStep, error) {
	return &bridgev2.LoginStep{
		Type:   bridgev2.LoginStepTypeUserInput,
		StepID: "com.github.robinchoice.mattermost.server_url",
		UserInputParams: &bridgev2.LoginUserInputParams{
			Fields: []bridgev2.LoginInputDataField{{
				Type: bridgev2.LoginInputFieldTypeURL,
				ID:   "server_url",
				Name: "Mattermost server URL",
			}},
		},
	}, nil
}

func (cl *CredentialsLogin) Cancel() {}

func (cl *CredentialsLogin) SubmitUserInput(ctx context.Context, input map[string]string) (*bridgev2.LoginStep, error) {
	if cl.client == nil {
		return cl.submitServerURL(ctx, input["server_url"])
	}
	return cl.submitCredentials(ctx, input["username"], input["password"])
}

func (cl *CredentialsLogin) submitServerURL(ctx context.Context, input string) (*bridgev2.LoginStep, error) {
	serverURL, err := mattermost.NormalizeServerURL(input)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}
	client := &mattermost.Client{ServerURL: serverURL, UserAgent: mautrix.DefaultUserAgent + " (login)"}
	if err = client.Ping(ctx); err != nil {
		return nil, bridgev2.RespError{
			ErrCode:    mautrix.MNotFound.ErrCode,
			Err:        fmt.Sprintf("No Mattermost server found at %s", serverURL),
			StatusCode: http.StatusBadRequest,
		}
	}
	cl.client = client
	return &bridgev2.LoginStep{
		Type:   bridgev2.LoginStepTypeUserInput,
		StepID: "com.github.robinchoice.mattermost.credentials",
		UserInputParams: &bridgev2.LoginUserInputParams{
			Fields: []bridgev2.LoginInputDataField{{
				Type: bridgev2.LoginInputFieldTypeUsername,
				ID:   "username",
				Name: "Username or email",
			}, {
				Type: bridgev2.LoginInputFieldTypePassword,
				ID:   "password",
				Name: "Password",
			}},
		},
	}, nil
}

func (cl *CredentialsLogin) submitCredentials(ctx context.Context, username, password string) (*bridgev2.LoginStep, error) {
	me, err := cl.client.Login(ctx, username, password)
	var apiErr *mattermost.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 {
		// Shown to the user as is, e.g. "Enter a valid email or username and/or password."
		return nil, bridgev2.RespError{
			ErrCode:    mautrix.MForbidden.ErrCode,
			Err:        apiErr.Message,
			StatusCode: http.StatusForbidden,
		}
	} else if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}
	serverURL := cl.client.ServerURL
	ul, err := cl.User.NewLogin(ctx, &database.UserLogin{
		ID:         makeUserLoginID(serverURL, me.ID),
		RemoteName: me.Username,
		RemoteProfile: status.RemoteProfile{
			Name:     me.DisplayName(),
			Username: me.Username,
			Email:    me.Email,
		},
		Metadata: &UserLoginMetadata{ServerURL: serverURL, Token: cl.client.Token},
	}, &bridgev2.NewLoginParams{DeleteOnConflict: true})
	if err != nil {
		return nil, err
	}
	go ul.Client.Connect(ul.Log.WithContext(cl.Main.Bridge.BackgroundCtx))
	return &bridgev2.LoginStep{
		Type:         bridgev2.LoginStepTypeComplete,
		StepID:       "com.github.robinchoice.mattermost.complete",
		Instructions: fmt.Sprintf("Logged in as %s on %s", me.Username, serverURL),
		CompleteParams: &bridgev2.LoginCompleteParams{
			UserLoginID: ul.ID,
			UserLogin:   ul,
		},
	}, nil
}

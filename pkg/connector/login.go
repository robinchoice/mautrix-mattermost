package connector

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/status"

	"github.com/robinchoice/mautrix-mattermost/pkg/mattermost"
)

const LoginFlowPassword = "password"

var loginFlows = []bridgev2.LoginFlow{{
	Name:        "Password",
	Description: "Log in with your Mattermost server URL, username or email, and password.",
	ID:          LoginFlowPassword,
}}

func (mc *MattermostConnector) GetLoginFlows() []bridgev2.LoginFlow {
	return loginFlows
}

func (mc *MattermostConnector) CreateLogin(ctx context.Context, user *bridgev2.User, flowID string) (bridgev2.LoginProcess, error) {
	if flowID != LoginFlowPassword {
		return nil, fmt.Errorf("unknown login flow %q", flowID)
	}
	return &PasswordLogin{User: user, Main: mc}, nil
}

type PasswordLogin struct {
	User *bridgev2.User
	Main *MattermostConnector
}

var _ bridgev2.LoginProcessUserInput = (*PasswordLogin)(nil)

func (pl *PasswordLogin) Start(ctx context.Context) (*bridgev2.LoginStep, error) {
	return &bridgev2.LoginStep{
		Type:   bridgev2.LoginStepTypeUserInput,
		StepID: "com.github.robinchoice.mattermost.password",
		UserInputParams: &bridgev2.LoginUserInputParams{
			Fields: []bridgev2.LoginInputDataField{{
				Type: bridgev2.LoginInputFieldTypeURL,
				ID:   "server_url",
				Name: "Mattermost server URL",
			}, {
				Type: bridgev2.LoginInputFieldTypeUsername,
				ID:   "login_id",
				Name: "Username or email",
			}, {
				Type: bridgev2.LoginInputFieldTypePassword,
				ID:   "password",
				Name: "Password",
			}},
		},
	}, nil
}

func (pl *PasswordLogin) Cancel() {}

func (pl *PasswordLogin) SubmitUserInput(ctx context.Context, input map[string]string) (*bridgev2.LoginStep, error) {
	serverURL, err := mattermost.NormalizeServerURL(input["server_url"])
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}
	cli := &mattermost.Client{ServerURL: serverURL, UserAgent: mautrix.DefaultUserAgent + " (login)"}
	me, err := cli.Login(ctx, input["login_id"], input["password"])
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
	ul, err := pl.User.NewLogin(ctx, &database.UserLogin{
		ID:         networkid.UserLoginID(me.ID),
		RemoteName: me.Username,
		RemoteProfile: status.RemoteProfile{
			Name:     me.DisplayName(),
			Username: me.Username,
			Email:    me.Email,
		},
		Metadata: &UserLoginMetadata{ServerURL: serverURL, Token: cli.Token},
	}, &bridgev2.NewLoginParams{DeleteOnConflict: true})
	if err != nil {
		return nil, err
	}
	go ul.Client.Connect(ul.Log.WithContext(pl.Main.Bridge.BackgroundCtx))
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

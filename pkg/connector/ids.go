package connector

import (
	"strings"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"
)

type UserLoginMetadata struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
}

func (mc *MattermostConnector) GetDBMetaTypes() database.MetaTypes {
	return database.MetaTypes{
		UserLogin: func() any {
			return &UserLoginMetadata{}
		},
	}
}

func (mc *MattermostClient) makeEventSender(userID string) bridgev2.EventSender {
	return bridgev2.EventSender{
		IsFromMe:    userID == mc.ownUserID,
		SenderLogin: makeUserLoginID(mc.Client.ServerURL, userID),
		Sender:      networkid.UserID(userID),
	}
}

func (mc *MattermostClient) makePortalKey(channelID string) networkid.PortalKey {
	pk := networkid.PortalKey{ID: networkid.PortalID(channelID)}
	if mc.UserLogin.Bridge.Config.SplitPortals {
		pk.Receiver = mc.UserLogin.ID
	}
	return pk
}

// makeUserLoginID puts the server in front of the user ID, so logins on two servers can't collide.
func makeUserLoginID(serverURL, userID string) networkid.UserLoginID {
	_, host, _ := strings.Cut(serverURL, "://")
	return networkid.UserLoginID(host + "/" + userID)
}

// parseUserLoginID returns the Mattermost user ID of a login. Logins created before
// the server was part of the ID are just the user ID.
func parseUserLoginID(loginID networkid.UserLoginID) string {
	id := string(loginID)
	return id[strings.LastIndexByte(id, '/')+1:]
}

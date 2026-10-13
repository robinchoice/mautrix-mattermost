package connector

import (
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
		SenderLogin: networkid.UserLoginID(userID),
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

package connector

import (
	"context"
	"strconv"

	"go.mau.fi/util/ptr"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/event"
)

func (mc *MattermostClient) GetChatInfo(ctx context.Context, portal *bridgev2.Portal) (*bridgev2.ChatInfo, error) {
	channel, err := mc.Client.GetChannel(ctx, string(portal.ID))
	if err != nil {
		return nil, err
	}
	topic := channel.Purpose
	if topic == "" {
		topic = channel.Header
	}
	return &bridgev2.ChatInfo{
		Name:  &channel.DisplayName,
		Topic: &topic,
		// Other members join as they post. Syncing the full member list comes later.
		Members: &bridgev2.ChatMemberList{
			MemberMap: map[networkid.UserID]bridgev2.ChatMember{
				networkid.UserID(mc.ownUserID): {
					EventSender: mc.makeEventSender(mc.ownUserID),
					Membership:  event.MembershipJoin,
				},
			},
		},
		Type: ptr.Ptr(database.RoomTypeDefault),
	}, nil
}

func (mc *MattermostClient) GetUserInfo(ctx context.Context, ghost *bridgev2.Ghost) (*bridgev2.UserInfo, error) {
	user, err := mc.Client.GetUser(ctx, string(ghost.ID))
	if err != nil {
		return nil, err
	}
	return &bridgev2.UserInfo{
		Identifiers: []string{"mattermost:" + user.Username},
		Name:        ptr.Ptr(user.DisplayName()),
		IsBot:       &user.IsBot,
		Avatar: &bridgev2.Avatar{
			ID: networkid.AvatarID(strconv.FormatInt(user.LastPictureUpdate, 10)),
			Get: func(ctx context.Context) ([]byte, error) {
				return mc.Client.GetProfileImage(ctx, user.ID)
			},
		},
	}, nil
}

func (mc *MattermostConnector) GetBridgeInfoVersion() (info, capabilities int) {
	return 1, 1
}

func (mc *MattermostConnector) GetCapabilities() *bridgev2.NetworkGeneralCapabilities {
	return &bridgev2.NetworkGeneralCapabilities{}
}

func (mc *MattermostClient) GetCapabilities(ctx context.Context, portal *bridgev2.Portal) *event.RoomFeatures {
	return &event.RoomFeatures{
		ID: "com.github.robinchoice.mattermost.capabilities.2026_10_13",
	}
}

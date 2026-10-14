package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/event"
)

func makeMatrixMessage(content *event.MessageEventContent, threadRoot *database.Message) *bridgev2.MatrixMessage {
	return &bridgev2.MatrixMessage{
		MatrixEventBase: bridgev2.MatrixEventBase[*event.MessageEventContent]{
			Content: content,
			Portal:  &bridgev2.Portal{Portal: &database.Portal{PortalKey: networkid.PortalKey{ID: "channel1"}}},
		},
		ThreadRoot: threadRoot,
	}
}

func TestConvertMatrixMessage(t *testing.T) {
	post, err := convertMatrixMessage(makeMatrixMessage(&event.MessageEventContent{
		MsgType:       event.MsgText,
		Body:          "hi there, see link",
		Format:        event.FormatHTML,
		FormattedBody: `<strong>hi</strong> <u>there</u>, see <a href="https://example.org">link</a>`,
	}, nil))
	require.NoError(t, err)
	assert.Equal(t, "channel1", post.ChannelID)
	assert.Equal(t, "**hi** there, see [link](https://example.org)", post.Message)
	assert.Empty(t, post.RootID)

	post, err = convertMatrixMessage(makeMatrixMessage(&event.MessageEventContent{
		MsgType: event.MsgNotice,
		Body:    "plain *as typed*",
	}, &database.Message{ID: "root1"}))
	require.NoError(t, err)
	assert.Equal(t, "plain *as typed*", post.Message)
	assert.Equal(t, "root1", post.RootID)

	// A Matrix thread started on a message that is itself a Mattermost thread reply
	post, err = convertMatrixMessage(makeMatrixMessage(&event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    "nested",
	}, &database.Message{ID: "reply1", ThreadRoot: "root1"}))
	require.NoError(t, err)
	assert.Equal(t, "root1", post.RootID)

	for _, msgType := range []event.MessageType{event.MsgImage, event.MsgFile, event.MsgEmote} {
		_, err = convertMatrixMessage(makeMatrixMessage(&event.MessageEventContent{MsgType: msgType, Body: "x"}, nil))
		assert.ErrorIs(t, err, bridgev2.ErrUnsupportedMessageType, msgType)
	}
}

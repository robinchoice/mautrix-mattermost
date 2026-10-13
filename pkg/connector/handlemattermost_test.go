package connector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maunium.net/go/mautrix/event"

	"github.com/robinchoice/mautrix-mattermost/pkg/mattermost"
)

func TestConvertPost(t *testing.T) {
	converted, err := convertPost(context.Background(), nil, nil, &mattermost.Post{Message: "**hi** there"})
	require.NoError(t, err)
	require.Len(t, converted.Parts, 1)
	content := converted.Parts[0].Content
	assert.Equal(t, event.MsgText, content.MsgType)
	assert.Equal(t, "**hi** there", content.Body)
	assert.Equal(t, "<strong>hi</strong> there", content.FormattedBody)
	assert.Nil(t, converted.ThreadRoot)

	converted, err = convertPost(context.Background(), nil, nil, &mattermost.Post{RootID: "root", FileIDs: []string{"f1"}})
	require.NoError(t, err)
	content = converted.Parts[0].Content
	assert.Equal(t, event.MsgNotice, content.MsgType)
	assert.Equal(t, "(attachment not bridged yet)", content.Body)
	require.NotNil(t, converted.ThreadRoot)
	assert.EqualValues(t, "root", *converted.ThreadRoot)
}

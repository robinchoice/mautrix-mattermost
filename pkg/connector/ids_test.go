package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserLoginID(t *testing.T) {
	id := makeUserLoginID("https://chat.example.org/mm", "u1")
	assert.EqualValues(t, "chat.example.org/mm/u1", id)
	assert.Equal(t, "u1", parseUserLoginID(id))
	assert.Equal(t, "u1", parseUserLoginID("u1"))
}

package mattermost

import (
	"encoding/json"
	"strings"
)

type User struct {
	ID                string `json:"id"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	Nickname          string `json:"nickname"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	LastPictureUpdate int64  `json:"last_picture_update"`
	IsBot             bool   `json:"is_bot"`
}

// DisplayName follows Mattermost's default "nickname, else full name, else username" order.
func (u *User) DisplayName() string {
	if u.Nickname != "" {
		return u.Nickname
	}
	if fullName := strings.TrimSpace(u.FirstName + " " + u.LastName); fullName != "" {
		return fullName
	}
	return u.Username
}

type ChannelType string

const (
	ChannelOpen    ChannelType = "O"
	ChannelPrivate ChannelType = "P"
	ChannelDirect  ChannelType = "D"
	ChannelGroup   ChannelType = "G"
)

type Channel struct {
	ID          string      `json:"id"`
	TeamID      string      `json:"team_id"`
	Type        ChannelType `json:"type"`
	DisplayName string      `json:"display_name"`
	Name        string      `json:"name"`
	Header      string      `json:"header"`
	Purpose     string      `json:"purpose"`
}

type Post struct {
	ID        string   `json:"id"`
	CreateAt  int64    `json:"create_at"`
	UserID    string   `json:"user_id"`
	ChannelID string   `json:"channel_id"`
	RootID    string   `json:"root_id"`
	Message   string   `json:"message"`
	Type      string   `json:"type"`
	FileIDs   []string `json:"file_ids"`
}

// IsSystem reports whether the post is a system message like "X joined the channel".
func (p *Post) IsSystem() bool {
	return p.Type != ""
}

const EventPosted = "posted"

type WebSocketEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
	Seq   int64           `json:"seq"`
}

type PostedData struct {
	ChannelType        ChannelType `json:"channel_type"`
	ChannelDisplayName string      `json:"channel_display_name"`
	SenderName         string      `json:"sender_name"`
	// Post is a JSON-encoded Post inside the JSON event.
	Post string `json:"post"`
}

func (d *PostedData) ParsePost() (*Post, error) {
	var post Post
	return &post, json.Unmarshal([]byte(d.Post), &post)
}

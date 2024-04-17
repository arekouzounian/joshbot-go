package main

type NewUserMessage struct {
	UserID        string `json:"userID"`        // Sender's user ID
	UnixTimestamp int64  `json:"unixTimestamp"` // Timestamp of the message
	JoshInt       uint8  `json:"joshInt"`       // 1 if it's 'josh' 0 otherwise
}

type JoshUpdateEvent struct {
	UserID   string `json:"userID"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

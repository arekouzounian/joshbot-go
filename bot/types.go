package main

import "github.com/bwmarrin/discordgo"

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

// stores tables for easy serialization/deserialization
type JoshCoinTableHolder struct {
	// userID to number of coins they earned today.
	DailyCoinsEarned map[string]int `json:"dailyCoinsEarned"`
	// userID to number of coins they earned before today
	CoinsBeforeToday map[string]int `json:"coinsBeforeToday"`
}

type JoshopItem struct {
	Name        string
	Description string
	Cost        int
	Button      *discordgo.Button
}

func NewJoshCoinTableHolder() *JoshCoinTableHolder {
	new := &JoshCoinTableHolder{
		DailyCoinsEarned: make(map[string]int),
		CoinsBeforeToday: make(map[string]int),
	}

	return new
}

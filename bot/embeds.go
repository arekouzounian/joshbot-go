package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func GenHelpCommandEmbed() *discordgo.MessageEmbed {
	fields := make([]*discordgo.MessageEmbedField, len(globalCommands))

	for i, cmd := range globalCommands {
		fields[i] = &discordgo.MessageEmbedField{
			Name:  cmd.Name,
			Value: cmd.Description,
		}
	}

	embed := discordgo.MessageEmbed{
		Title: "joshbot commands",
		Description: `pay attention, josh. these are the commands you can use.
		
Make sure you always use commands in the DM. If you use it on the server it will count towards a non-josh. josh`,
		Fields: fields,
	}

	return &embed
}

// Generates an embed with the user's josh coin information
func GenJoshCoinCommandEmbed(userID string) *discordgo.MessageEmbed {
	if TableHolder == nil {
		return nil
	}

	coinsEarnedBeforeToday := 0
	if cbt, exists := TableHolder.CoinsBeforeToday[userID]; exists {
		coinsEarnedBeforeToday = cbt // not the fun kind :(
	} else {
		TableHolder.CoinsBeforeToday[userID] = 0
	}
	beforeTodayField := &discordgo.MessageEmbedField{
		Name:   "Josh Coins: Before Today",
		Value:  fmt.Sprintf("You have earned `%d` josh coins before today", coinsEarnedBeforeToday),
		Inline: false,
	}

	coinsEarnedToday := 0
	if ct, exists := TableHolder.DailyCoinsEarned[userID]; exists {
		coinsEarnedToday = ct
	} else {
		TableHolder.DailyCoinsEarned[userID] = 0
	}
	earnedTodayField := &discordgo.MessageEmbedField{
		Name:   "Josh Coins: Today",
		Value:  fmt.Sprintf("You have earned `%d` josh coins today", coinsEarnedToday),
		Inline: false,
	}

	sumField := &discordgo.MessageEmbedField{
		Name:   "Josh Coins: Total",
		Value:  fmt.Sprintf("Total: `%d` josh coins", coinsEarnedBeforeToday+coinsEarnedToday),
		Inline: false,
	}

	fields := []*discordgo.MessageEmbedField{beforeTodayField, earnedTodayField, sumField}

	embed := discordgo.MessageEmbed{
		Title:       "Josh Coin",
		Description: "josh",
		Fields:      fields,
	}

	return &embed
}

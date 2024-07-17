package main

import (
	"fmt"
	"os"

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

	coinsEarnedBeforeToday := GetCoinsBeforeToday(userID)
	beforeTodayField := &discordgo.MessageEmbedField{
		Name:   "Josh Coins: Before Today",
		Value:  fmt.Sprintf("You have earned `%d` josh coins before today", coinsEarnedBeforeToday),
		Inline: false,
	}

	coinsEarnedToday := GetDailyCoins(userID)
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

func GenAnnouncementEmbed(input_file string) (*discordgo.MessageEmbed, error) {
	f, err := os.ReadFile(input_file)
	if err != nil {
		return nil, err
	}

	embed := discordgo.MessageEmbed{
		Title:       "Announcement",
		Description: string(f),
	}

	return &embed, nil
}

func GenJoshopEmbed(userID string) *discordgo.MessageEmbed {

	fields := make([]*discordgo.MessageEmbedField, len(JoshopItems))
	for i, x := range JoshopItems {
		fields[i] = &discordgo.MessageEmbedField{
			Name:  x.Name,
			Value: x.Description,
		}
	}

	embed := discordgo.MessageEmbed{
		Title:       "Joshop",
		Description: fmt.Sprintf("*Current Balance: %d*", GetTotalCoins(userID)),
		Fields:      fields,
	}

	return &embed
}

package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

// generates josh log and user table
// BUG: joshlog stored in reverse chronological order
// UPDATE: this is in fact not a bug and will improve performance
func GenTables(session *discordgo.Session, joshlog_output_file string, usertable_output_file string) {
	// first generate josh log
	var userJoshCount map[string]uint = make(map[string]uint)
	var userNonJoshCount map[string]uint = make(map[string]uint)

	messages, err := session.ChannelMessages(JOSH_CHANNEL_ID, 100, "", "", "")
	if err != nil {
		fmt.Printf("Error grabbing messages: %s\n", err.Error())
		return
	}

	file, err := os.OpenFile(joshlog_output_file, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Error opening file %s: %s\n", joshlog_output_file, err.Error())
		return
	}
	defer file.Close()

	// api schema is unix_timestamp,user_id,joshInt
	for len(messages) > 0 {
		last_id := messages[len(messages)-1].ID

		for _, message := range messages {
			var joshInt uint8

			if message.Content != "josh" {
				joshInt = 0
				userNonJoshCount[message.Author.ID] += 1
			} else {
				joshInt = 1
				userJoshCount[message.Author.ID] += 1
			}

			_, err := file.WriteString(fmt.Sprintf("%d,%s,%d\n", message.Timestamp.Unix(), message.Author.ID, joshInt))
			if err != nil {
				fmt.Printf("Error writing to %s: %s", joshlog_output_file, err.Error())
			}
		}

		messages, err = session.ChannelMessages(JOSH_CHANNEL_ID, 100, last_id, "", "")
		if err != nil {
			fmt.Printf("Error grabbing earlier messages: %s\n", err.Error())
			break
		}
	}

	//now grab user table
	userfile, err := os.OpenFile(usertable_output_file, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Error opening file %s: %s\n", usertable_output_file, err.Error())
		return
	}
	defer userfile.Close()

	users, err := session.GuildMembers(GUILD_ID, "", 1000)
	if err != nil {
		fmt.Printf("Error getting users: %s\n", err.Error())
		return
	}

	// schema is id,name,avatar,josh,nonjosh
	for _, user := range users {
		if user.User.Bot {
			continue
		}

		_, err := userfile.WriteString(fmt.Sprintf("%s,%s,%s,%d,%d\n", user.User.ID, user.User.Username, user.AvatarURL(""), userJoshCount[user.User.ID], userNonJoshCount[user.User.ID]))
		if err != nil {
			fmt.Printf("Error writing to %s: %s", usertable_output_file, err.Error())
		}
	}

	fmt.Println("Successful Scrape.")
}

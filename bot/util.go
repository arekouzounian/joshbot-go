package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Sends a DM to the Josh of the Week.
// Any failures will be logged, but won't be fatal.
func dmJoshOtw(session *discordgo.Session) {
	resp, err := http.Get(API_URL + JOSH_OTW_ENDPOINT)
	if err != nil {
		log.Printf("Error getting josh of the week: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		log.Printf("Couldn't read request response: %s", err.Error())
	}

	fields := strings.Split(buf.String(), ",")

	channel, err := session.UserChannelCreate(fields[0])
	if err != nil {
		log.Printf("Error creating DM channel with user %s: %s", fields[1], err.Error())
		return
	}

	_, err = session.ChannelMessageSend(channel.ID, "congratulations josh, you are now this week's josh of the week. http://joshbot.xyz")
	if err != nil {
		log.Printf("Error DM'ing user: %s", err.Error())
		return
	}

	log.Printf("Sent congratulatory message to user %s", fields[1])
}

// Checks if every user is named 'josh'
// Can be expensive and might be rate limited for large, uninitialized servers
func checkUsernames(session *discordgo.Session) {
	users, err := session.GuildMembers(GUILD_ID, "", 1000)
	if err != nil {
		log.Printf("Error retrieving user list during name check: %s", err.Error())
		return
	}

	for _, user := range users {
		if user.Nick != "josh" {
			err := session.GuildMemberNickname(GUILD_ID, user.User.ID, "josh")
			if err != nil {
				log.Printf("Error changing nickname of user %s to josh: %s", user.User.Username, err.Error())
			}
		}
	}
}

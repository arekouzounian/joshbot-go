package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// event handler for edited messages
func messageUpdate(session *discordgo.Session, message *discordgo.MessageUpdate) {
	if !DebugMode && message.GuildID != GUILD_ID {
		return
	}

	if message.Author.System {
		return
	}

	if message.Content != "josh" {
		err := session.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			fmt.Printf("Error deleting edited message: %s", err.Error())
		}
	}

}

// event handler for whenever a message is sent
func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Checking if message should be ignored
	if !DebugMode && message.GuildID != GUILD_ID {
		return
	}

	channel, err := session.Channel(message.ChannelID)
	if err != nil {
		log.Printf("Error getting channel: %s. The error might be related to a thread deletion event.", err.Error())
		return
	}

	if channel.Type == discordgo.ChannelTypeDM {
		return
	}

	if channel.IsThread() {
		_, err := session.ChannelDelete(channel.ID)
		if err != nil {
			log.Printf("Error deleting thread channel: %s", err.Error())
		}
		return
	}

	if message.Author.System {
		return
	}

	// Message fits scope; form API request
	reqData := NewUserMessage{
		UserID:        message.Author.ID,
		UnixTimestamp: time.Now().Unix(),
		JoshInt:       0,
	}

	if message.Content != "josh" { // non-josh message

		log.Printf("Non-josh message detected: %s: %s\n", message.Author.Username, message.Content)
		DeleteMsg(session, message.ChannelID, message.ID)

	} else { // josh message

		if LastMsg == nil {
			panic("LastMsg nil, unable to check for double-josh")
		}

		time_gap := time.Since(LastMsg.Timestamp)
		if LastMsg.Author.ID == message.Author.ID && time_gap < (time.Hour*DOUBLE_JOSH_SPAN) {
			log.Printf("User %s sent a double-josh", message.Author.Username)
			DeleteMsg(session, message.ChannelID, message.ID)
			DMUser(session, message.Author.ID, fmt.Sprintf("Double-josh detected: do not double josh within the span of %d hours", DOUBLE_JOSH_SPAN))
		} else {
			// valid josh
			reqData.JoshInt = 1
			LastMsg = message.Message
		}
	}

	// Marshal data for request body
	json, err := json.Marshal(reqData)
	if err != nil {
		log.Printf("Error marshalling json data: %s", err.Error())
		return
	}

	// Debug Mode will ignore API requests
	if !DebugMode {
		resp, err := http.Post(API_URL+NEW_MSG_ENDPOINT, "application/json", bytes.NewBuffer(json))
		if err != nil {
			log.Printf("Error creating api request: %s", err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			buf := new(strings.Builder)
			_, err := io.Copy(buf, resp.Body)
			if err != nil {
				log.Printf("Couldn't read request response: %s", err.Error())
			}

			log.Printf("Status code %d, server responded with: %s", resp.StatusCode, buf.String())
		} else {
			log.Printf("Success: %s's josh event sent to API", message.Author.Username)
		}
	}
}

// For our purposes, this event requires the same work as the user update event.
// This function will simply call the userUpdate()
func userJoin(session *discordgo.Session, newUser *discordgo.GuildMemberAdd) {
	cast := discordgo.GuildMemberUpdate{
		Member: newUser.Member,
	}

	userUpdate(session, &cast)
}

// When a user update event is generated, some information about them is changed/new
// We make sure they have the right name/role, then push changes to the API
func userUpdate(session *discordgo.Session, update *discordgo.GuildMemberUpdate) {
	if !DebugMode && update.GuildID != GUILD_ID {
		return
	}

	// change their name to josh
	// give them the josh role
	if update.Nick != "josh" {
		log.Printf("Non-josh nickname detected on user %s", update.User.Username)
		err := session.GuildMemberNickname(update.GuildID, update.User.ID, "josh")
		if err != nil {
			log.Printf("Error assigning josh nickname: %s", err.Error())
		} else {
			log.Printf("Josh nickname established on %s successfully", update.User.Username)
		}
	}
	hasJosh := false
	for _, roleID := range update.Roles {
		if roleID == JOSH_ROLE_ID {
			hasJosh = true
		}
	}
	if !hasJosh {
		log.Printf("Josh role not assigned to user %s", update.User.Username)
		err := session.GuildMemberRoleAdd(update.GuildID, update.User.ID, JOSH_ROLE_ID)
		if err != nil {
			log.Printf("Error assigning josh role: %s", err.Error())
		} else {
			log.Printf("Josh role assigned to user %s successfully", update.User.Username)
		}
	}

	// send api request
	if !DebugMode {

		reqData := JoshUpdateEvent{
			UserID:   update.User.ID,
			Username: update.User.Username,
			Avatar:   update.AvatarURL(""),
		}
		json, err := json.Marshal(reqData)
		if err != nil {
			log.Printf("Error marshalling json: %s", err.Error())
			return
		}

		resp, err := http.Post(API_URL+ADD_USER_ENDPOINT, "application/json", bytes.NewBuffer(json))
		if err != nil {
			log.Printf("Error creating API request: %s", err.Error())
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			buf := new(strings.Builder)
			_, err := io.Copy(buf, resp.Body)
			if err != nil {
				log.Printf("Couldn't read request response: %s", err.Error())
			}

			log.Printf("Status code %d, server responded with: %s", resp.StatusCode, buf.String())
		} else {
			log.Printf("Success: %s's guild update registered with API", update.User.Username)
		}
	}
}

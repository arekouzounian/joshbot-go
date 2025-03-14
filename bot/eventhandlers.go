package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	PERCENT_CHANCE_TO_RESPOND = 20
)

// event handler for edited messages
func MessageUpdate(session *discordgo.Session, message *discordgo.MessageUpdate) {
	if !DebugMode && message.GuildID != GUILD_ID {
		return
	}

	if message.Author.System || message.Author.Bot {
		return
	}

	if message.Content != "josh" {
		log.Println("Detected updated non-josh; deleting.")

		err := session.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			log.Printf("Error deleting edited message: %s", err.Error())
		}
	}

}

// event handler for whenever a message is sent
func MessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Checking if message should be ignored

	// first check if it is a DM
	channel, err := session.Channel(message.ChannelID)
	if err != nil {
		log.Printf("Error getting channel: %s. The error might be related to a thread deletion event.\n", err.Error())
		return
	}

	if channel.Type == discordgo.ChannelTypeDM {
		roll := (rand.Int() % 100) + 1

		if roll < PERCENT_CHANCE_TO_RESPOND {
			log.Printf("Attempting to respond to %s's DM\n", message.Author.Username)
			if err = sendUserRandomGif(session, message.Author.ID); err != nil {
				log.Printf("Sending user random GIF failed: %v\n", err)
			}
		}

		return
	}

	// then check for the right guild
	if (!DebugMode && message.GuildID != GUILD_ID) || message.Author.System {
		return
	}

	// if it is the right guild but its a thread then delete
	if channel.IsThread() {
		_, err := session.ChannelDelete(channel.ID)
		if err != nil {
			log.Printf("Error deleting thread channel: %s\n", err.Error())
		}
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
			JoshCoinGenerateCheck(session, message.Message)
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
func UserJoin(session *discordgo.Session, newUser *discordgo.GuildMemberAdd) {
	cast := discordgo.GuildMemberUpdate{
		Member: newUser.Member,
	}

	UserUpdate(session, &cast)
}

// When a user update event is generated, some information about them is changed/new
// We make sure they have the right name/role, then push changes to the API
func UserUpdate(session *discordgo.Session, update *discordgo.GuildMemberUpdate) {
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

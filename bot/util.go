package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
)

// Initializes any variables needed for later use; intended to be called shortly after session opens.
// Any error returned should be a fatal error.
func InitializeState(session *discordgo.Session) error {

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Printf("Error creating scheduler: %s", err.Error())
		return err
	}
	scheduler.Start()
	defer scheduler.Shutdown()

	// Job to DM the josh of the week every week
	_, err = scheduler.NewJob(
		gocron.CronJob(
			"0 2 * * 0",
			false,
		),
		gocron.NewTask(
			dmJoshOtw,
			session,
		),
	)
	if err != nil {
		log.Printf("Error scheduling job: %s", err.Error())
		return err
	}
	log.Println("Created josh of the week scheduler successfully.")

	msg, err := session.ChannelMessages(JOSH_CHANNEL_ID, 1, "", "", "")
	if err != nil {
		log.Printf("Error getting last message: %s", err.Error())
		return err
	}
	LastMsg = msg[0]

	checkUsernames(session)

	inputID := ""
	if SlashCommandDebug {
		inputID = "779964589768179742"
	}

	err = UpdateAndRegisterGlobalCommands(session, inputID)
	if err != nil {
		log.Printf("Error updating and/or registering slash commands: %s", err.Error())
		return err
	}

	log.Println("Slash commands registered and operational.")

	return nil
}

func DeleteMsg(session *discordgo.Session, channelID string, messageID string) {
	err := session.ChannelMessageDelete(channelID, messageID)
	if err != nil {
		log.Printf("Error deleting message: %s", err.Error())

		perms, err := session.State.UserChannelPermissions(session.State.User.ID, channelID)
		if err != nil {
			log.Printf("Unable to retreive user channel permissions: %s", err.Error())
		}

		if perms&discordgo.PermissionAdministrator == 0 {
			log.Printf("Not running as administrator on this channel! Won't be able to remove messages.")
			session.ChannelMessageSend(channelID, "I need to be admin to work effectively.")
		}

	} else {
		log.Printf("Message deleted successfully.")
	}
}

func DMUser(session *discordgo.Session, userID string, message string) error {
	channel, err := session.UserChannelCreate(userID)
	if err != nil {
		log.Printf("Error creating DM channel with user %s: %s", userID, err.Error())
		return err
	}

	_, err = session.ChannelMessageSend(channel.ID, message)
	if err != nil {
		log.Printf("Error DM'ing user: %s", err.Error())
	}

	return err
}

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

	err = DMUser(session, fields[0], "congratulations josh, you are now this week's josh of the week. http://joshbot.xyz")

	if err == nil {
		log.Printf("Sent congratulatory message to user %s", fields[1])
	}
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

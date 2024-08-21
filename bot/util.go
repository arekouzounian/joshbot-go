package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
)

// Initializes any variables needed for later use; intended to be called shortly after session opens.
// Returns true if successful, false if there were fatal errors
func InitializeState(session *discordgo.Session) bool {
	// setting up scheduler
	// can't use implicit declaration or else 'Scheduler' will not be stored globally
	var err error
	Scheduler, err = gocron.NewScheduler()
	if err != nil {
		log.Printf("Error creating scheduler: %s", err.Error())
		return false
	}
	Scheduler.Start()

	err = scheduleJobs(session)
	if err != nil {
		log.Printf("Error scheduling job(s): %s", err.Error())
		return false
	}

	// setting last message to check for double josh
	msg, err := session.ChannelMessages(JOSH_CHANNEL_ID, 1, "", "", "")
	if err != nil {
		log.Printf("Error getting last message: %s", err.Error())
		return false
	}
	LastMsg = msg[0]

	// checks to see if users are named josh
	checkUsernames(session)

	// slash command initialization
	inputID := ""
	if SlashCommandDebug {
		inputID = "779964589768179742"
	}
	err = UpdateAndRegisterGlobalCommands(session, inputID)
	if err != nil {
		log.Printf("Error updating and/or registering slash commands: %s", err.Error())
		return false
	}
	log.Println("Slash commands registered and operational.")

	// Init TableHolder
	if _, err := os.Stat(JOSHCOIN_FILE_DEFAULT); os.IsNotExist(err) {
		TableHolder = NewJoshCoinTableHolder()
	} else {
		err := DeserializeTablesFromFile(JOSHCOIN_FILE_DEFAULT)
		if err != nil {
			log.Printf("Error getting tables from file: %s", err.Error())
			panic(err.Error())
		}
	}

	return true
}

// Does any tasks that need to be done at exit
// Should theoretically catch any panics
func AtExit() {
	err := SerializeTablesToFile(JOSHCOIN_FILE_DEFAULT)
	if err != nil {
		log.Printf("Error serializig tables to file: %s", err.Error())
	}

	err = Scheduler.Shutdown()
	if err != nil {
		log.Printf("Error shutting down scheduler: %s", err.Error())
	}
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
	if userID == session.State.User.ID {
		return nil
	}

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

func DMUserEmbed(session *discordgo.Session, userID string, embed *discordgo.MessageEmbed) error {
	if userID == session.State.User.ID {
		return nil
	}

	channel, err := session.UserChannelCreate(userID)
	if err != nil {
		log.Printf("Error creating DM channel with user %s: %s", userID, err.Error())
		return err
	}

	_, err = session.ChannelMessageSendEmbed(channel.ID, embed)
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

	var respBodyData [][]string

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&respBodyData)
	if err != nil {
		log.Printf("Error decoding json body data: %s", err.Error())
		return
	}

	fields := respBodyData[0]

	err = DMUser(session, fields[0], "congratulations josh, you are now this week's josh of the week. http://joshbot.xyz")

	if err == nil {
		log.Printf("Sent congratulatory message to user %s", fields[1])
	} else {
		log.Printf("Error sending dm to user %s: %s", fields[1], err.Error())
	}
}

// Backs up the josh coin tables to the serialization file
// updates internal TableHolder, migrating daily coins over to coins before today
func joshCoinDailyReset(session *discordgo.Session) {
	for userID, coins := range TableHolder.DailyCoinsEarned {
		TableHolder.CoinsBeforeToday[userID] += coins
		TableHolder.DailyCoinsEarned[userID] = 0
	}

	err := SerializeTablesToFile(JOSHCOIN_FILE_DEFAULT)
	if err != nil {
		log.Printf("Error serializing tables on daily reset: %s", err.Error())
	} else {
		log.Printf("Successful daily table reset")
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

// function that schedules all jobs; broken into own function for cleanliness
func scheduleJobs(session *discordgo.Session) error {
	// Job to DM the josh of the week every week
	_, err := Scheduler.NewJob(
		gocron.CronJob(
			"0 2 * * 1",
			false,
		),
		gocron.NewTask(
			dmJoshOtw,
			session,
		),
	)
	if err != nil {
		return err
	}
	log.Println("Created josh of the week scheduler successfully.")

	_, err = Scheduler.NewJob(
		gocron.CronJob(
			"0 0 * * *",
			false,
		),
		gocron.NewTask(
			joshCoinDailyReset,
			session,
		),
	)
	if err != nil {
		return err
	}
	log.Println("Created josh coin daily reset job successfully.")

	return nil
}

func sendUserRandomGif(session *discordgo.Session, userID string) error {
	const TENOR_BASE_URL = "https://tenor.googleapis.com/v2/search?q=josh"

	key, exists := os.LookupEnv("TENOR_API_KEY")
	if !exists {
		log.Println("No TENOR_API_KEY environment variable found; unable to send random gifs to users.")
		return nil
	}

	url := TENOR_BASE_URL + "&key=" + key + "&limit=1&random=true"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Error in making HTTP request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	var apiResp TenorApiResponse
	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&apiResp); err != nil {
		return fmt.Errorf("Error in decoding JSON: %v", err)
	}

	return DMUser(session, userID, apiResp.Results[0].MediaFormats.Gif.URL)
}

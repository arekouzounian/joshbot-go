package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
)

/*

	TODO:
	- strip perms
	- check for thread creation (?)
	- delete edited messages

*/

var (
	Token     string
	DebugMode bool
	LogFile   string
)

// hardcoded server ID; allows testing on other server
// might be a better way to do this
const (
	GUILD_ID          = "715798257661509743"
	JOSH_ROLE_ID      = "716065561385238589"
	JOSH_CHANNEL_ID   = "715798258190123031"
	API_URL           = "http://joshbot.xyz:6969"
	ADD_USER_ENDPOINT = "/api/v1/joshupdate"
	NEW_MSG_ENDPOINT  = "/api/v1/newjosh"
	JOSH_OTW_ENDPOINT = "/api/v1/joshotw"
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.BoolVar(&DebugMode, "d", false, "sets the bot to debug mode")
	flag.StringVar(&LogFile, "o", "./joshbot.log", "The file to output logs to. By default, creates a file in the current directory named 'joshbot.log'")
	flag.Parse()
}

func main() {
	if Token == "" {
		fmt.Println("Flag missing! Must specify bot token with -t flag.")
		return
	}

	file, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening logfile: %s could not be opened.", LogFile)
		return
	}
	defer file.Close()

	log.SetOutput(file)

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Error creating discord session: %s\n", err.Error())
	}

	log.Println("Discord session started")

	// https://discord.com/developers/docs/topics/gateway#gateway-intents
	dg.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuildMembers | discordgo.IntentsDirectMessages

	dg.AddHandler(messageCreate)
	dg.AddHandler(userJoin)
	dg.AddHandler(userUpdate)
	dg.AddHandler(messageUpdate)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening discord connection: %s", err.Error())
	}
	defer dg.Close()

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("Error creating scheduler: %s", err.Error())
	}
	scheduler.Start()
	defer scheduler.Shutdown()

	_, err = scheduler.NewJob(
		gocron.CronJob(
			"0 2 * * 0",
			false,
		),
		gocron.NewTask(
			dmJoshOtw,
			dg,
		),
	)
	if err != nil {
		log.Fatalf("Error scheduling job: %s", err.Error())
	}

	fmt.Println("Created josh of the week scheduler successfully.")

	if DebugMode {
		fmt.Println("WARNING: Debug mode activated. Server access not restricted, API requests not being made.")

		// trickery
		// err := dg.GuildMemberRoleRemove(GUILD_ID, "392796102132367364", "715798870256386131")
		// if err != nil {
		// 	fmt.Println(err.Error())
		// }
	}
	fmt.Println("Bot running! Use Ctrl-C to exit.")
	checkUsernames(dg)
	// GenTables(dg, "../api/joshlog.csv", "../api/users.csv")
	sigchannel := make(chan os.Signal, 1)
	signal.Notify(sigchannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigchannel

	log.Printf("Received interrupt, shutting down\n\n")
}

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

	_, err = session.ChannelMessageSend(channel.ID, "congratulations josh, you are now this week's josh of the week. https://joshbot.xyz")
	if err != nil {
		log.Printf("Error DM'ing user: %s", err.Error())
		return
	}

	log.Printf("Sent congratulatory message to user %s", fields[1])
}

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

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if !DebugMode && message.GuildID != GUILD_ID {
		return
	}

	channel, err := session.Channel(message.ChannelID)
	if err != nil {
		log.Printf("Error getting channel: %s", err.Error())
		log.Printf("The error might be related to a thread deletion event")
		return
	}

	if channel.Type == discordgo.ChannelTypeDM {
		return
	}

	if channel.IsThread() {
		_, err := session.ChannelDelete(channel.ID)
		if err != nil {
			log.Fatalf("Error deleting thread channel: %s", err.Error())
		}
		return
	}

	if message.Author.System {
		return
	}

	reqData := NewUserMessage{
		UserID:        message.Author.ID,
		UnixTimestamp: time.Now().Unix(),
		JoshInt:       0,
	}

	if message.Content != "josh" {
		log.Printf("Non-josh message detected: %s: %s\n", message.Author.Username, message.Content)
		err := session.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			log.Printf("Error deleting message: %s", err.Error())

			perms, err := session.State.UserChannelPermissions(session.State.User.ID, message.ChannelID)
			if err != nil {
				log.Fatalf("Unable to retreive user channel permissions: %s", err.Error())
			}

			if perms&discordgo.PermissionAdministrator == 0 {
				log.Printf("Not running as administrator on this channel! Won't be able to remove messages.")
				session.ChannelMessageSend(message.ChannelID, "I need to be admin to work effectively.")
			}

		} else {
			log.Printf("Message deleted successfully.")
		}
	} else {
		reqData.JoshInt = 1
	}

	json, err := json.Marshal(reqData)
	if err != nil {
		log.Printf("Error marshalling json data: %s", err.Error())
		return
	}

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

// When a user joins, they will be assigned the josh role and this information will be
// communicated to the API.
func userJoin(session *discordgo.Session, newUser *discordgo.GuildMemberAdd) {
	cast := discordgo.GuildMemberUpdate{
		Member: newUser.Member,
	}

	userUpdate(session, &cast)
}

// server handles duplicates so we can just call the above
func userUpdate(session *discordgo.Session, update *discordgo.GuildMemberUpdate) {
	// change their name to josh
	// give them the josh role

	if !DebugMode && update.GuildID != GUILD_ID {
		return
	}

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

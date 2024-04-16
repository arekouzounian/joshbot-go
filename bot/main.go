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
)

var (
	Token   string
	LogFile string
)

// hardcoded server ID; allows testing on other server
// might be a better way to do this
const (
	GuildID         = "715798257661509743"
	ApiURL          = "http://localhost:5000"
	AddUserEndpoint = "/api/v1/joshjoin"
	NewMsgEndpoint  = "/api/v1/newjosh"
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
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
	dg.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuildMembers

	dg.AddHandler(messageCreate)
	dg.AddHandler(userJoin)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening discord connection: %s", err.Error())
	}

	fmt.Println("Bot running! Use Ctrl-C to exit.")
	sigchannel := make(chan os.Signal, 1)
	signal.Notify(sigchannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigchannel

	dg.Close()

	log.Printf("Received interrupt, shutting down\n\n")
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	reqData := NewUserMessage{
		UserID:        message.Author.ID,
		UnixTimestamp: time.Now().Unix(),
		JoshInt:       0,
	}

	if strings.ToLower(message.Content) != "josh" {
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

	resp, err := http.Post(ApiURL+NewMsgEndpoint, "application/json", bytes.NewBuffer(json))
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

// When a user joins, they will be assigned the josh role and this information will be
// communicated to the API.
func userJoin(session *discordgo.Session, newUser *discordgo.GuildMemberAdd) {

}

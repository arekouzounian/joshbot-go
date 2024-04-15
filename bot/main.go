package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token   string
	LogFile string
)

// hardcoded server ID; allows testing on other server
// might be a better way to do this
const (
	GuildID = "715798257661509743"
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
	if message.GuildID != GuildID {
		return
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

		// non josh processing

	} else {
		// handle josh message !

		/*
			will send a POST request to an internal api about new josh message

			features of api:
			- POST: new josh msg
			- POST: update member list (member left, member joined)
			- POST: non-joshes deleted
			- GET: josh of the week
		*/
	}
}

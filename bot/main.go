package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token        string
	DebugMode    bool
	GenTableMode bool
	LogFile      string
	LastMsg      *discordgo.Message
)

// hardcoded server ID; allows testing on other server
// might be a better way to do this
// if the bot is always going to be on the same server as api,
// can change API_URL -> http://localhost:6969
const (
	DOUBLE_JOSH_SPAN  = 12
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
	flag.BoolVar(&GenTableMode, "g", false, "will use the bot to generate user table and josh log table, then exits.")
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

	if GenTableMode {
		fmt.Printf("Entering table generator mode...")
		err = GenTables(dg, "./joshlog.csv", "./users.csv")
		if err == nil {
			fmt.Println("Tables generated successfully at './joshlog.csv', './users.csv'")
		}
		return
	}

	if DebugMode {
		fmt.Println("WARNING: Debug mode activated. Server access not restricted, API requests not being made.")
	}

	err = InitializeState(dg)
	if err != nil {
		return
	}

	fmt.Println("Bot startup successful! Use Ctrl-C to exit.")

	sigchannel := make(chan os.Signal, 1)
	signal.Notify(sigchannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigchannel

	log.Printf("Received interrupt, shutting down\n\n")
}

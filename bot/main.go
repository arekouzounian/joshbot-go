package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
)

var (
	Token             string
	DebugMode         bool
	SlashCommandDebug bool
	AnnouncementMode  bool
	AnnouncementFile  string
	GenTableMode      bool
	RmCmdMode         bool
	LogFile           string
	LastMsg           *discordgo.Message
	Scheduler         gocron.Scheduler
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
	API_URL           = "http://localhost:8080"
	ADD_USER_ENDPOINT = "/api/v2/joshupdate"
	NEW_MSG_ENDPOINT  = "/api/v2/newjosh"
	JOSH_OTW_ENDPOINT = "/api/v2/joshotw"
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.BoolVar(&DebugMode, "d", false, "sets the bot to debug mode")
	flag.BoolVar(&SlashCommandDebug, "sd", false, "sets the bot to slash command debug mode")
	flag.BoolVar(&GenTableMode, "gentable", false, "will use the bot to generate user table and josh log table, then exits.")
	flag.BoolVar(&RmCmdMode, "rmcmd", false, "will delete all registered slash commands, then exits.")
	flag.BoolVar(&AnnouncementMode, "announcement", false, "will take an announcement msg from an input file, then dm that announcement to every user")
	flag.StringVar(&AnnouncementFile, "af", "", "the target file to use as an announcement message")
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
		log.Printf("Fatal: error opening logfile: %s could not be opened.", LogFile)
		return
	}
	defer file.Close()

	log.SetOutput(file)

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Printf("Error creating discord session: %s\n", err.Error())
		return
	}

	log.Println("Discord session started")
	defer log.Printf("Discord session ended\n\n")

	// https://discord.com/developers/docs/topics/gateway#gateway-intents
	dg.Identify.Intents = discordgo.IntentGuildMessages | discordgo.IntentGuildMembers | discordgo.IntentsDirectMessages

	dg.AddHandler(messageCreate)
	dg.AddHandler(userJoin)
	dg.AddHandler(userUpdate)
	dg.AddHandler(messageUpdate)
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var name string
		if i.Type == discordgo.InteractionApplicationCommand {
			name = i.ApplicationCommandData().Name
		} else {
			name = i.MessageComponentData().CustomID
		}

		if handler, exists := interactionHandlers[name]; exists {
			handler(s, i)
		} else {
			log.Printf("No handler exists for interaction %s", name)
		}
	})

	err = dg.Open()
	if err != nil {
		log.Printf("Error opening discord connection: %s", err.Error())
		return
	}
	defer dg.Close()

	if GenTableMode {
		fmt.Printf("Entering table generator mode...")
		err = GenTables(dg, "./joshlog.csv", "./users.csv")
		if err == nil {
			fmt.Println("Tables generated successfully at './joshlog.csv', './users.csv'")
		} // error printing already done internally
		return
	} else if RmCmdMode {
		fmt.Println("Entering command cleanup mode...")
		err = CleanupGlobalCommands(dg)
		if err == nil {
			fmt.Println("Commands deleted successfully.")
		} else {
			fmt.Printf("Error deleting application command(s): %s\n", err.Error())
		}
		return
	} else if AnnouncementMode {
		fmt.Println("Entering Announcement mode...")
		if AnnouncementFile == "" {
			fmt.Println("You need to specify the target file with the -af flag.")
			return
		}

		embed, err := GenAnnouncementEmbed(AnnouncementFile)
		if err != nil {
			fmt.Printf("Error generating announcement embed: %s\n", err.Error())
			return
		}

		users, err := dg.GuildMembers(GUILD_ID, "", 1000)
		if err != nil {
			fmt.Printf("Error getting guild members: %s\n", err.Error())
			return
		}

		for _, user := range users {
			if user.User.Bot {
				continue
			}

			// for testing
			// if user.User.ID != "MY USER ID FOR TESTING" {
			// 	continue
			// }

			err := DMUserEmbed(dg, user.User.ID, embed)
			if err != nil {
				return
			}
			fmt.Printf("Sent announcement to user %s successfully.", user.User.Username)
		}

		fmt.Println("Announcements sent successfully.")
		return
	}

	if DebugMode {
		fmt.Println("WARNING: Debug mode activated. Server access not restricted, API requests not being made.")
	} else if SlashCommandDebug {
		fmt.Println("WARNING: Slash Command Debug mode activated. Message create hook not being checked.")
	}

	defer AtExit() //needs to be deferred to catch a panic

	success := InitializeState(dg)
	if !success {
		return
	}

	fmt.Println("Bot startup successful! Use Ctrl-C to exit.")

	sigchannel := make(chan os.Signal, 1)
	signal.Notify(sigchannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sigchannel

	log.Printf("Received interrupt, shutting down\n")
}

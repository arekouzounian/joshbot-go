package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	dmPermission = true

	globalCommands = []*discordgo.ApplicationCommand{
		{
			Name:         "help",
			Description:  "this command shows how to use joshbot",
			Type:         discordgo.ChatApplicationCommand,
			DMPermission: &dmPermission,
		},
		{
			Name:         "joshcoin",
			Description:  "shows you how many josh coins you have",
			Type:         discordgo.ChatApplicationCommand,
			DMPermission: &dmPermission,
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{GenHelpCommandEmbed()},
				},
			})

			postCommandLogging(i, "help", err)
		},
		"joshcoin": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var userID string
			if i.User == nil {
				userID = i.Member.User.ID
			} else {
				userID = i.User.ID
			}

			embed := GenJoshCoinCommandEmbed(userID)
			var err error
			if embed != nil {
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{GenJoshCoinCommandEmbed(userID)},
					},
				})
			} else {
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "No tables exist!!",
					},
				})
			}

			postCommandLogging(i, "joshcoin", err)
		},
	}
)

func postCommandLogging(i *discordgo.InteractionCreate, command_name string, err error) {
	if err != nil {
		log.Printf("Error responding to '%s' command: %s", command_name, err.Error())
	} else {
		if i.User != nil {
			log.Printf("Command '%s' executed by user %s in DM successfully", command_name, i.User.Username)
		} else {
			log.Printf("Command '%s' executed by user %s in channel successfully", command_name, i.Member.User.Username)
		}
	}
}

func CleanupGlobalCommands(session *discordgo.Session) error {
	registered, err := session.ApplicationCommands(session.State.User.ID, "")
	if err != nil {
		return err
	}

	for _, cmd := range registered {
		err = session.ApplicationCommandDelete(session.State.User.ID, "", cmd.ID)

		if err != nil {
			return err
		}

		if RmCmdMode {
			fmt.Printf("Deleted command %s successfully\n", cmd.Name)
		}
	}

	return nil
}

// checks if local commands match what discord has registered, then registers whatever is missing
// shouldn't run too often; just at startup is enough
func UpdateAndRegisterGlobalCommands(session *discordgo.Session, guildID string) error {
	registered, err := session.ApplicationCommands(session.State.User.ID, guildID)
	if err != nil {
		return err
	}

	// maps the command pointer to a string representing its ID
	// if the ID is empty it needs to be registered, otherwise it needs to be updated
	need_update := make(map[*discordgo.ApplicationCommand]string)

	for _, local_command := range globalCommands {
		found := false

		if SlashCommandDebug {
			fmt.Printf("Checking slash command '%s' for registration\n", local_command.Name)
		}

		for _, registered_command := range registered {
			if SlashCommandCheckEqual(registered_command, local_command) {
				// command exists as is in discord api
				found = true
				if SlashCommandDebug {
					fmt.Printf("Slash command '%s' already registered and unmodified\n", local_command.Name)
				}

			} else if local_command.Name == registered_command.Name {
				// command exists but local version differs
				log.Printf("Found differing slash command: %s", local_command.Name)
				need_update[local_command] = registered_command.ID
				found = true

				if SlashCommandDebug {
					fmt.Printf("Slash command '%s' already registered but needs modification\n", local_command.Name)
				}
			}
		}

		if !found {
			log.Printf("Found unregistered slash command: %s", local_command.Name)
			need_update[local_command] = ""

			if SlashCommandDebug {
				fmt.Printf("Slash command '%s' needs registration\n", local_command.Name)
			}
		}
	}

	for cmd, id := range need_update {
		if id == "" {
			_, err := session.ApplicationCommandCreate(session.State.User.ID, guildID, cmd)
			if err != nil {
				return err
			}
			log.Printf("Successfully registered slash command: %s", cmd.Name)
		} else {
			_, err := session.ApplicationCommandEdit(session.State.User.ID, guildID, id, cmd)
			if err != nil {
				return err
			}
			log.Printf("Successfully updated slash command: %s", cmd.Name)
		}
	}

	return nil
}

func RegisterGlobalCommand(session *discordgo.Session, cmd *discordgo.ApplicationCommand) {
	session.ApplicationCommandCreate(session.State.User.ID, "", cmd)
}

func SlashCommandCheckEqual(a *discordgo.ApplicationCommand, b *discordgo.ApplicationCommand) bool {
	if a == nil || b == nil {
		return false
	}

	var equal bool = true

	equal = equal && (a.Name == b.Name)
	equal = equal && (a.Description == b.Description)
	equal = equal && (a.Type == b.Type)

	// got rid of dm permission check hope I don't need it later
	equal = equal && (len(a.Options) == len(b.Options))

	for _, a_option := range a.Options {
		found := false
		for _, b_option := range b.Options {
			if a_option.Name == b_option.Name {
				found = true

				equal = equal && (a_option.Description == b_option.Description)
				equal = equal && (a_option.Type == b_option.Type)
				// more checks later if needed
			}
		}

		equal = equal && found
	}

	return equal
}

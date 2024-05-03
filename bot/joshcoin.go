package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"

	"github.com/bwmarrin/discordgo"
)

const (
	JOSHCOIN_CHANCE_DEFAULT = 5
	DAILY_MAX               = 3
	JOSHCOIN_FILE_DEFAULT   = "./joshcointables.json"
)

var (
	TableHolder *JoshCoinTableHolder
)

// checks if a josh coin is generated on a message
// will DM the user that they have received a josh coin
func JoshCoinGenerateCheck(session *discordgo.Session, message *discordgo.Message) {
	// generates in [1, 100]
	roll := (rand.Int() % 100) + 1

	if roll <= JOSHCOIN_CHANCE_DEFAULT && TableHolder.DailyCoinsEarned[message.Author.ID] < DAILY_MAX {
		// they got a josh coin
		log.Printf("User %s got a josh coin\n", message.Author.Username)
		TableHolder.DailyCoinsEarned[message.Author.ID] += 1
		DMUser(session, message.Author.ID, "josh, you just earned a josh coin")
	}
}

// Deserializes the tables from the given file into the TableHolder variable
// input_file: the file to read from. default "./joshcointables.json"
func DeserializeTablesFromFile(input_file string) error {
	if input_file == "" {
		input_file = JOSHCOIN_FILE_DEFAULT
	}

	b, err := os.ReadFile(input_file)
	if err != nil {
		return err
	}

	tablesFromFile := &JoshCoinTableHolder{}

	err = json.Unmarshal(b, tablesFromFile)
	if err != nil {
		return err
	}

	TableHolder = tablesFromFile

	return nil
}

// Serializes the tables from TableHolder into the given file
// output_file: the file to save JSON to. default "./joshcointables.json"
func SerializeTablesToFile(output_file string) error {
	if output_file == "" {
		output_file = JOSHCOIN_FILE_DEFAULT
	}

	b, err := json.Marshal(*TableHolder)
	if err != nil {
		return err
	}

	err = os.WriteFile(output_file, b, 0666)
	if err != nil {
		return err
	}

	return nil
}

/*

Naive Approach:
- store cached copy in memory as obj
- whenever cached copy is updated then serialize changes to file
	- a bit costly for cases w/ a large amount of users
- on startup, get cached copy by deserializing file

- Advantages:
	- simpler to implement
	- makes it easy to deal with daily coin reset
- Disadvantages:
	- performance is slow
	- doesn't scale well with high volume of messages or users

Alternative:
- store cached copy in memory as obj
- only serialize to file on bot shutdown
	- OR upon daily coin amount reset (?)
	- theoretically not even necessary to do it on daily reset, but it's better to periodically "back up"
- on startup, get cached copy by deserializing file

- pretty much opposite advantages/disadvantages as the naive approach


Idea:
- users can claim one coin per day
- users have a chance get coins from messages
- users have a daily limit of how many coins they can earn per day

- daily coins tracked in struct (id->coins earned that day)
- every day at midnight that struct is reset and changes are pushed to the other struct (and serialized depending on serialization schema)

*/

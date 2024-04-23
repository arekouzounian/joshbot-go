# Joshbot 
This folder contains all the code for the discord bot this project is built on. 

### Relevant files
- [`./main.go`](./main.go) contains the entrypoint for the executable.
- [`hack.go`](./hack.go) has hacky utility functions like 'GenTables', which will generate user tables and joshlog tables.
- [`eventhandlers.go`](./eventhandlers.go) holds functions for that can be hooked into the discord session to handle various events (message create, user update, etc.)
- [`types.go`](./types.go) is the file that holds type definitions (structs)


### Future additions 

#### Josh Coin system
- every user has a chance to earn josh coins 
- Josh coins earned via a daily slash command redeem (most likely in the bot's DM)
- There is a daily limit on the number of josh coins one can get per day (3 joins?)
- The first time the josh of the week redeems for the week, they get a 5-join bonus 

#### Josh Coin Economy 
- example reward: double josh. This allows you to send a double josh without your second message being deleted as per the no double josh policy. (low cost)
- example reward: josh power. Gives you a chance (josh percent) to have each individual josh message be worth one more josh point. The chance itself can be another reward itself (very high cost)
- example reward: josh percent. Increases the chance that your josh message will give you two josh points. (medium cost)
- example rewards: hsoj. Reduces another user's josh score by hsoj power, min. 1. (low cost)
- example rewards: hsoj power. Increases the amount of josh's removed from a user when using hsoj. (high cost)
/*
	By Narwhal Prime

	initial code base and tutorial:
	https://github.com/bwmarrin/discordgo/tree/master/examples/pingpong
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	Email    string
	Password string
	Token    string
	BotID    string
)

type UserData struct {
	DiscordID	string
	Username	string
	PADID		int
	Description	string
}

func init() {
	flag.StringVar(&Email, "e", "", "Account Email")
	flag.StringVar(&Password, "p", "", "Account Password")
	flag.StringVar(&Token, "t", "", "Account Token")
	flag.Parse()
}

const (
	EMPTY_PADID = 0
	PROFILES_DIR = "profiles/"
)

func main() {

	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New(Email, Password, Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// Store the account ID for later use.
	BotID = u.ID

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}

// Load user's data from file
// TODO datastore will change to a database later
func loadData(id string) (*UserData, error) {
	
	// load the file data
	body, err := ioutil.ReadFile(PROFILES_DIR + id)
	if err != nil {
		return nil, err
	}

	// parse the body data
	lines := strings.Split(string(body), "\n")
	username := lines[0]
	padid, _ := strconv.Atoi(lines[1])
	description := lines[2]

	return &UserData{DiscordID: id, Username: username, PADID: padid, Description: description}, nil
}

// Save user's data to file
func saveData(data *UserData) (error) {

	// check if file for this Discord ID exists already
	username := data.Username
	
	currData, err := loadData(data.DiscordID)
	var padid int
	var description string

	// get the existing data, if any
	if err == nil {
		padid = currData.PADID
		description = currData.Description
		fmt.Printf("currData contents = %s %i %s\n", currData.Username, currData.PADID, currData.Description)
	} else {
		padid = EMPTY_PADID
		description = ""
	}

	// prepare to update existing data if necessary
	if data.PADID != EMPTY_PADID {
		padid = data.PADID
	}
	if data.Description != "" {
		description = data.Description
	}
	
	body := username + "\n" + strconv.Itoa(padid) + "\n" + description
	return ioutil.WriteFile(PROFILES_DIR + data.DiscordID, []byte(body), 0600)
}

func getGroup(padid int) (string, error) {
	groupDigit := (padid / int(math.Pow(10, 6))) % 10
	switch groupDigit {
		case 0, 5:	return "A", nil
		case 1, 6:	return "B", nil
		case 2, 7:	return "C", nil
		case 3, 8:	return "D", nil
		case 4, 9:	return "E", nil
	}
	return "", errors.New("Group not able to be identified")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(sess *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == BotID {
		return
	}

	discordID := m.Author.ID
	username := m.Author.Username

	if strings.HasPrefix(m.Content, "-lampad") {

		// read in a series of tokens
		message := "Hi there! I'm the LamPAD Bot.\n" +
			"Here are my commands: myid, mygroup, mydescription, schedule, getinfo, ping, pong\n" + 
			"Example: -lampad ping"
		inputTokens := strings.Split(m.Content, " ")
		if len(inputTokens) >= 2 {

			firstTok := inputTokens[1]
			switch firstTok {

				case "ping":
					message = "Pong! :)"
				case "pong":
					message = "Ping! :D"
				case "myid":
					if len(inputTokens) == 2 {
						data, err := loadData(discordID)

						// error indicates user not found
						if err == nil {
							message = username + ", your PAD ID is " + strconv.Itoa(data.PADID)
						} else {
							message = "I currently don't have your PAD ID yet, " + username + ".\n" + 
								"Use \"-lampad myid #########\" (without quotes) to record or update your PAD ID!"
						}

					} else {
						
						inputPADID, errAtoi := strconv.Atoi(inputTokens[2])
						if errAtoi == nil {
							newData := &UserData{DiscordID: discordID, Username: username, PADID: inputPADID, Description: ""}
							// fmt.Printf("id username padid = %s %i %s", discordID, username, inputPADID)
							err := saveData(newData)

							// determine what group based on PAD ID
							group, _ := getGroup(inputPADID)

							if err == nil {
								message = "I have updated your PAD ID, " + username + "! You are in group " + group
							} else {
								message = "Oops! Something went wrong and I wasn't able to update your PAD ID.\n" + 
									"Please let Narwhal Prime know what happened!"
							}
						} else {
							message = "Hmm... that doesn't appear to be a valid ID."
						}
					}
					
				case "mygroup":
					data, err := loadData(discordID)
					if err == nil {
						group, err2 := getGroup(data.PADID)
						if err2 == nil {
							message = username + ", you are in group " + group
						} else {
							message = "Oops! Something went wrong and I wasn't able to determine your group.\n" + 
								"Please let Narwhal Prime know what happened!"	
						}
					} else {
						message = "I currently don't have your PAD ID yet and thus don't know your group, " + username + ".\n" + 
								"Use \"-lampad myid #########\" (without quotes) to record or update your PAD ID!"
					}

				case "mydescription":
					if len(inputTokens) == 2 { // get own description
						data, err := loadData(discordID)

						// error indicates user not found
						if err == nil {
							message = username + ": \"" + data.Description + "\""
						} else {
							message = username + ", you can use \"-lampad myid Awesome text here\" (without quotes) to record/update a description of yourself.\n" +
								"Try putting in things like your favorite leads or next dungeon/farm goals!"
						}

					} else { // update own description
						newDescription := strings.Join(inputTokens[2:], " ")

						newData := &UserData{DiscordID: discordID, Username: username, PADID: EMPTY_PADID, Description: newDescription}
						err := saveData(newData)

						if err == nil {
							message = username + ", your description has been updated"
						} else {
							message = "Oops! Something went wrong and I wasn't able to update your description.\n" + 
								"Please let Narwhal Prime know what happened!"
						}
					}
				case "getinfo":
					message = "This command is a work-in-progress; check back later!"
				case "schedule":
					message = "This command is a work-in-progress; check back later!"
				default:
					message = "Hello! Please specify a valid command.\n" + 
						"My commands: myid, mygroup, schedule, ping, pong"
			}
		}

		// send appropriate message back to the text channel
		_, _ = sess.ChannelMessageSend(m.ChannelID, message)
	}
}

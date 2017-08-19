package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

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
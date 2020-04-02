package main

import (
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/pojntfx/nextcloud-talk-jitsi-bot/pkg/v2/pkg/client"
)

func main() {
	var (
		url        = os.Getenv("NEXTCLOUD_URL")
		username   = os.Getenv("NEXTCLOUD_USERNAME")
		password   = os.Getenv("NEXTCLOUD_PASSWORD")
		dbLocation = os.Getenv("DB_LOCATION")
		jitsiURL   = os.Getenv("JITSI_URL")
	)

	chatChan, statusChan := make(chan client.Chat), make(chan string)
	bot := client.NewNextcloudTalk(url, username, password, dbLocation, chatChan, statusChan)

	defer bot.Close()
	if err := bot.Open(); err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := bot.ReadRooms(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := bot.ReadChats(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		for status := range statusChan {
			log.Println(status)
		}
	}()

	for chat := range chatChan {
		log.Printf(`Received message from "%v" ("%v") in room "%v" with ID "%v": "%v"`, chat.ActorDisplayName, chat.ActorID, chat.Token, chat.ID, chat.Message)

		reg := regexp.MustCompile("^#video(chat|call)")

		if reg.Match([]byte(chat.Message)) {
			log.Printf(`"%v" ("%v") has requested a video call in room "%v" with ID "%v"; creating video call.`, chat.ActorDisplayName, chat.ActorID, chat.Token, chat.ID)

			bot.CreateChat(chat.Token, fmt.Sprintf("@%v started a video call. Tap on %v to join!", chat.ActorID, jitsiURL+"/"+chat.Token))
		}
	}
}

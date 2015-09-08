package main

import (
	"github.com/dghubble/oauth1"
	"github.com/mitchellh/go-homedir"
	"github.com/yuntan/tw/go-tw"

	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var (
	settings    Settings
	token       *oauth1.Token
	lastReplyID string
	lastTweetID []string
)

func main() {
	// load settings
	dir, _ := homedir.Expand(SETTING_DIR)
	b, err := ioutil.ReadFile(dir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("config file (%s) not found.", SETTING_DIR)
		} else {
			log.Println(err)
		}
		os.Exit(1)
	}

	if err := json.Unmarshal(b, &settings); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// show event info
	log.Println("registered events:")
	for _, ev := range settings.Events {
		log.Printf("%d %s\n", ev.Hour, ev.Message)
	}
	log.Println()

	token = &oauth1.Token{
		Token:       settings.AccessToken,
		TokenSecret: settings.AccessTokenSecret,
	}

	for { // main loop
		checkReply()
		processEvents()
		time.Sleep(time.Duration(settings.CheckInterval) * time.Minute)
	}
}

func checkReply() {
	log.Println("checking replies...")

	resp, err := tw.GetMentions(token)
	if err != nil {
		log.Println(err)
		return
	}
	if resp.StatusCode/100 != 2 {
		log.Println("Status: ", resp.StatusCode)
		return
	}

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	var data []map[string]interface{}
	dec.Decode(&data) // TODO error
	id := data[0]["id_str"].(string)
	name := data[0]["user"].(map[string]interface{})["screen_name"].(string)
	text := data[0]["text"].(string)
	at, _ := time.Parse(time.RubyDate, data[0]["created_at"].(string))

	if id == lastReplyID {
		return
	}
	lastReplyID = id

	log.Println("Reply received:")
	log.Printf("\"%s\" from @%s at %s\n", text, name, at)

	if name != settings.TargetUser {
		return
	}

	now := time.Now()
	for _, ev := range settings.Events {
		switch {
		case now.Hour() < ev.Hour:
			ev.next = now.Truncate(24 * time.Hour).Add(time.Duration(ev.Hour) * time.Hour)

		case now.Hour() > ev.Hour:
			ev.next = now.Truncate(24 * time.Hour).Add(time.Duration(24+ev.Hour) * time.Hour)

		default:
			for _, sub := range settings.ReplyOK {
				if strings.Contains(text, sub) {
					log.Printf("OK for %d %s\n", ev.Hour, ev.Message)
					// TODO save log
					ev.next = now.Truncate(time.Hour).Add(24 * time.Hour)
					log.Println("next nofify time: ", ev.next.Format(time.Stamp))
					return
				}
			}
			for _, sub := range settings.ReplyNG {
				if strings.Contains(text, sub) {
					log.Printf("NG for %d %s\n", ev.Hour, ev.Message)
					// TODO save log
					ev.next = now.Truncate(time.Hour).Add(24 * time.Hour)
					log.Println("next nofify time: ", ev.next.Format(time.Stamp))
					return
				}
			}

			ev.next = now.Add(time.Duration(settings.NotifyInterval) * time.Minute)
		}
	}
}

func processEvents() {
	for _, ev := range settings.Events {
		now := time.Now()
		if now.Before(ev.next) {
			continue
		}

		log.Printf("processing evet %d %s...\n", ev.Hour, ev.Message)
		tweet := strings.Replace(settings.Format, "{message}", ev.Message, 1)
		tweet = strings.Replace(tweet, "{time}", now.Format(time.Kitchen), 1)
		resp, err := tw.Tweet(tweet, token)
		if err != nil {
			log.Println(err)
			return
		}
		if resp.StatusCode/100 != 2 {
			log.Println("status: ", resp.StatusCode)
			return
		}

		dec := json.NewDecoder(resp.Body)
		defer resp.Body.Close()
		var data map[string]interface{}
		dec.Decode(&data) // TODO error
		id := data["id_str"].(string)
		lastTweetID = append(lastTweetID, id)
	}
}

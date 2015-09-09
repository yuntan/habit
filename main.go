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
			log.Fatalf("config file (%s) not found.", SETTING_DIR)
		} else {
			log.Fatalln(err)
		}
	}

	if err := json.Unmarshal(b, &settings); err != nil {
		log.Fatalln(err)
	}

	// show event info
	log.Println("registered events:")
	for _, hb := range settings.Habits {
		log.Printf("%d %s\n", hb.Hour, hb.Message)
	}
	log.Println()

	token = &oauth1.Token{
		Token:       settings.AccessToken,
		TokenSecret: settings.AccessTokenSecret,
	}

	select {
	case <-time.Tick(time.Hour):
		for _, habit := range settings.Habits {
			if time.Now().Hour() == habit.Hour {
				go processHabit(habit)
			}
		}
	}
}

func processHabit(habit Habit) {
	count := 0
	ok := false

	select {
	case <-time.Tick(time.Duration(settings.CheckInterval) * time.Minute):
		if res, suc := checkReply(habit); suc {
			ok = res
			break
		}

	case <-time.Tick(time.Duration(settings.NotifyInterval) * time.Minute):
		if count > settings.NotifyCount {
			break
		}
		notify(habit)
		count++
	}

	_ = ok // TODO save log
}

func checkReply(habit Habit) (result, success bool) {
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
	dec.Decode(&data)
	if len(data) == 0 {
		return
	}
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

	for _, sub := range settings.ReplyOK {
		if strings.Contains(text, sub) {
			log.Printf("OK for %d %s\n", habit.Hour, habit.Message)
			success = true
			result = true
			return
		}
	}
	for _, sub := range settings.ReplyNG {
		if strings.Contains(text, sub) {
			log.Printf("NG for %d %s\n", habit.Hour, habit.Message)
			success = true
			result = false
			return
		}
	}
	return
}

func notify(habit Habit) (id string, success bool) {
	log.Printf("notify %d %s...\n", habit.Hour, habit.Message)
	tweet := strings.Replace(settings.Format, "{message}", habit.Message, 1)
	tweet = strings.Replace(tweet, "{time}", time.Now().Format(time.Kitchen), 1)
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
	id = data["id_str"].(string)
	return
}

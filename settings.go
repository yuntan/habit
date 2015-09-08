package main

import "time"

const SETTING_DIR = "~/.habit.json"

type Settings struct {
	Events            []Event  `json:"events"`
	NotifyInterval    int      `json:"notify_interval"` // min
	NotifyCount       int      `json:"notify_count"`
	CheckInterval     int      `json:"check_interval"` // min
	Format            string   `json:"format"`
	ReplyOK           []string `json:"reply_ok"`
	ReplyNG           []string `json:"reply_ng"`
	AccessToken       string   `json:"access_token"`
	AccessTokenSecret string   `json:"access_token_secret"`
	TargetUser        string   `json:"target_user"`
}

type Event struct {
	Hour    int    `json:"hour"` // 24 tense
	Message string `json:"message"`
	next    time.Time
}

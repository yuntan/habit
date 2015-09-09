package main

const SETTING_DIR = "~/.habit.json"

type Settings struct {
	Habits            []Habit  `json:"habits"`
	NotifyInterval    int      `json:"notify_interval"` // min
	NotifyCount       int      `json:"notify_count"`
	CheckInterval     int      `json:"check_interval"` // min
	Format            string   `json:"format"`
	ReplyOK           []string `json:"reply_ok"`
	ReplyNG           []string `json:"reply_ng"`
	TargetUser        string   `json:"target_user"`
	AccessToken       string   `json:"access_token"`
	AccessTokenSecret string   `json:"access_token_secret"`
}

type Habit struct {
	Hour    int    `json:"hour"` // 24 tense
	Message string `json:"message"`
}

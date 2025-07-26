package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type UserPreferences struct {
	Port      int      `json:"port"`
	Users     []User   `json:"users"`
	SudoAuth  string   `json:"sudo-auth"`
	Blacklist []string `json:"ip-blacklist"`
	Whitelist []string `json:"ip-whitelist"`
	Services  []string `json:"services"`
	YrFiles   struct {
		SavePath string `json:"save-path"`
	} `json:"yrFiles"`
	YrSound struct {
		SavePath string `json:"save-path"`
	} `json:"yrSound"`
}

func get_datentime() string {
	var Time time.Time = time.Now()
	return Time.Format("2006-01-02 15:04:05")
}
func get_settings() UserPreferences {
	var settings UserPreferences
	var config_json, _ = os.Open(filepath.Join(filePath, "config.json"))
	defer config_json.Close()

	json.NewDecoder(config_json).Decode(&settings)
	return settings
}

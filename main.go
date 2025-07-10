package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

var filePath string

// gui dirs

//go:embed gui/yrFiles/*
var yrFiles embed.FS
//go:embed gui/yrPics/*
var yrPics embed.FS

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type UserPreferences struct {
	Port      int             														`json:"port"`
	Users     []User          														`json:"users"`
	SudoAuth  string          														`json:"sudo-auth"`
	Blacklist []string        														`json:"ip-blacklist"`
	Whitelist []string        														`json:"ip-whitelist"`
	Services  []string        														`json:"services"`
	YrFiles   struct {SavePath string `json:"save-path"`} `json:"yrFiles"`
	YrPics    struct {SavePath string `json:"save-path"`} `json:"yrPics"`
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
func check_allowed(r *http.Request) bool {
	var host, _, _ = net.SplitHostPort(r.RemoteAddr)
	if slices.Contains(get_settings().Blacklist, host) {
		return false
	}
	if whitelist := get_settings().Whitelist; len(whitelist) > 0 {
		var allowed bool = false
		if slices.Contains(whitelist, host) {
			allowed = true
		}
		if !allowed {
			return false
		}
	}

	return true
}
func check_auth(r *http.Request) bool {
	if len(get_settings().Users) == 0 {
		return true
	}
	var username string = r.URL.Query().Get("username")
	var password string = r.URL.Query().Get("password")

	for _, user := range get_settings().Users {
		if !(user.Username == username) || !(user.Password == password) {
			return false
		}
	}

	return true
}

func main() {
	filePath, _ = filepath.Abs(".")

	for _, user := range get_settings().Users {
		os.Mkdir(filepath.Join(save_path, user.Username), 0755)
	}

	fmt.Println(BLUE + "Initializing " + PINK + "listener..." + RESET)
	var listener, err = net.Listen("tcp", ":"+strconv.Itoa(get_settings().Port))
	if err != nil {
		fmt.Println(RED+"ERROR while initializing listener:", err.Error())
		os.Exit(1)
	} else {
		fmt.Println(GREEN + "Listener created!" + RESET + " Listening at " + PINK + listener.Addr().String() + RESET)
	}

	fmt.Println(BLUE + "Starting " + PINK + "HTTP server..." + RESET)

	// misc
	http.HandleFunc("/users-qm", func(w http.ResponseWriter, r *http.Request) {
		if len(get_settings().Users) == 0 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		log(ATTEMPT, "CLIENT: Does this server have users?", true)
	})
	http.HandleFunc("/isUser-qm", func(w http.ResponseWriter, r *http.Request) {
		if check_auth(r) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
		log(ATTEMPT, "CLIENT: Are these dudes with you?", true)
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// yrFiles
	if slices.Contains(get_settings().Services, "files") {
		http.HandleFunc("/yrFiles/", func(w http.ResponseWriter, r *http.Request) {
			if !check_allowed(r) {
				return
			}
			var yrFilesSub, _ = fs.Sub(yrFiles, "gui/yrFiles")
			http.StripPrefix("/yrFiles/", http.FileServer(http.FS(yrFilesSub))).ServeHTTP(w, r)
		})
		http.HandleFunc("/delete-file", http_deleteFile)
		http.HandleFunc("/get-files", http_getFiles)
		http.HandleFunc("/get-content", http_getContent)
		http.HandleFunc("/rename-file", http_renameFile)
		http.HandleFunc("/upload-chunk", http_uploadChunk)
	}
	if slices.Contains(get_settings().Services, "pics") {
		http.HandleFunc("/yrPics/", func(w http.ResponseWriter, r *http.Request) {
			if !check_allowed(r) {
				return
			}
			var yrPicsSub, _ = fs.Sub(yrPics, "gui/yrPics")
			http.StripPrefix("/yrPics/", http.FileServer(http.FS(yrPicsSub))).ServeHTTP(w, r)
		})
	}

	// log for start
	fmt.Print("\033c")
	fmt.Printf(GREEN+"Server initiated!"+RESET+" Serving at "+PINK+"%s"+RESET+"\n", listener.Addr())
	fmt.Println("Press " + PINK + "<Ctrl-C>" + RESET + " to stop the server.")
	fmt.Println(PINK + "-------------------------------------------------" + RESET)

	http.Serve(listener, nil)
}

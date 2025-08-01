package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var filePath string

// gui dirs

//go:embed gui/yrFiles/*
var yrFiles embed.FS

//go:embed gui/yrSound/*
var yrSound embed.FS

// settings types
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
		if (user.Username == username) && (user.Password == password) {
			return true
		}
	}

	return false
}

func main() {
	filePath, _ = filepath.Abs(".")

	log(ATTEMPT, "Initializing services for `yrHost`...", false)
	for _, user := range get_settings().Users {
		if slices.Contains(get_settings().Services, "files") {
			log(STEP, "Preparing `yrFiles`...\n", false)
			os.Mkdir(filepath.Join(yf_savePath, user.Username), 0755)
		}
		if slices.Contains(get_settings().Services, "sound") {
			log(STEP, "Preparing `yrSound`...\n", false)
			os.Mkdir(filepath.Join(ys_savePath, user.Username), 0755)
			os.Mkdir(filepath.Join(ys_savePath, user.Username, "picture"), 0755)
			var songs []Song
			var err error = Index(filepath.Join(ys_savePath, user.Username), &songs)
			if err != nil {
				log(ERROR, "An error occured while trying to index all songs for `yrSound`.", true)
				return
			}
			var ysDB, _ = sql.Open("sqlite3", filepath.Join(ys_savePath, user.Username, ".db"))
			defer ysDB.Close()
			ysDB.Exec(`
				drop table if exists songs;
				create table if not exists songs (
					id INTEGER PRIMARY KEY,
					filepath TEXT,
					title TEXT,
					artist TEXT,
					album TEXT,
					albumArtist TEXT,
					year INTEGER,
					track INTEGER,
					disc INTEGER
				);
			`)
			for _, song := range songs {
				ysDB.Exec(`
					insert or ignore into songs (
						filepath, title, artist, album, albumArtist, year, track, disc
					) values (?, ?, ?, ?, ?, ?, ?, ?)
				`,
					song.FilePath, song.Title, song.Artist, song.Album,
					song.AlbumArtist, song.Year, song.Track,
					song.Disc,
				)
			}
		}
	}

	log(ATTEMPT, "Initializing listener...", false)
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
	// yrSound
	if slices.Contains(get_settings().Services, "sound") {
		http.HandleFunc("/yrSound/", func(w http.ResponseWriter, r *http.Request) {
			if !check_allowed(r) {
				return
			}
			var yrSoundSub, _ = fs.Sub(yrSound, "gui/yrSound")
			http.StripPrefix("/yrSound/", http.FileServer(http.FS(yrSoundSub))).ServeHTTP(w, r)
		})
		http.HandleFunc("/get-cover", http_getCover)
		http.HandleFunc("/get-id", http_getID)
		http.HandleFunc("/get-albums", http_getAlbums)
		http.HandleFunc("/get-album", http_getAlbum)
		http.HandleFunc("/get-songs", http_getSongs)
		http.HandleFunc("/get-artists", http_getArtists)
		http.HandleFunc("/get-song-info", http_getSongInfo)
		http.HandleFunc("/get-song-blob", http_getSongBlob)
	}

	// log for start
	fmt.Print("\033c")
	fmt.Printf(GREEN+"Server initiated!"+RESET+" Serving at "+PINK+"%s"+RESET+"\n", listener.Addr())
	fmt.Println("Press " + PINK + "<Ctrl-C>" + RESET + " to stop the server.")
	fmt.Println(PINK + "-------------------------------------------------" + RESET)

	http.Serve(listener, nil)
}

package main

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	_ "github.com/mattn/go-sqlite3"
)

var filePath string

// gui dirs

//go:embed gui/yrFiles/*
var yrFiles embed.FS

//go:embed gui/yrText/*
var yrText embed.FS

//go:embed gui/yrSound/*
var yrSound embed.FS

// settings types
func check_allowed(c *fiber.Ctx) bool {
	var host string = c.IP()
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
func check_auth(c *fiber.Ctx) bool {
	if len(get_settings().Users) == 0 {
		return true
	}
	var username string = c.Query("username")
	var password string = c.Query("password")

	for _, user := range get_settings().Users {
		if (user.Username == username) && (user.Password == password) {
			return true
		}
	}

	return false
}

func main() {
	filePath, _ = filepath.Abs(".")
	//unknown ext by golang (those little shits)

	mime.AddExtensionType(".mov", "video/quicktime")
	mime.AddExtensionType(".heic", "image/heic")
	mime.AddExtensionType(".heif", "image/heif")
	mime.AddExtensionType(".mkv", "video/x-matroska")
	mime.AddExtensionType(".webm", "video/webm")
	mime.AddExtensionType(".flac", "audio/flac")
	mime.AddExtensionType(".opus", "audio/opus")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".epub", "application/epub+zip")
	mime.AddExtensionType(".ics", "text/calendar")
	mime.AddExtensionType(".md", "text/markdown")

	log(ATTEMPT, "Initializing services for `yrHost`...", false)
	for _, user := range get_settings().Users {
		if slices.Contains(get_settings().Services, "files") {
			log(STEP, "Preparing `yrFiles`...\n", false)
			os.Mkdir(filepath.Join(yf_savePath, user.Username), 0755)
		}
		if slices.Contains(get_settings().Services, "sound") {
			log(STEP, "Preparing `yrSound`...\n", false)
			os.Mkdir(filepath.Join(ys_savePath, user.Username), 0755)
			os.Mkdir(filepath.Join(ys_savePath, user.Username, "pictures"), 0755)
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
					song.AlbumArtist, song.Year, song.Track, song.Disc,
				)
			}
		}
	}

	log(ATTEMPT, "Initializing listener...", false)
	var listener, err = net.Listen("tcp", ":"+strconv.Itoa(get_settings().Port))
	var app *fiber.App = fiber.New(fiber.Config{DisableStartupMessage: true})
	if err != nil {
		fmt.Println(RED+"ERROR while initializing listener:", err.Error())
		os.Exit(1)
	} else {
		fmt.Println(GREEN + "Listener created!" + RESET + " Listening at " + PINK + listener.Addr().String() + RESET)
	}

	fmt.Println(BLUE + "Starting " + PINK + "HTTP server..." + RESET)

	// misc
	app.Get("/users-qm", func(c *fiber.Ctx) error {
		log(ATTEMPT, "CLIENT: Does this server have users?", true)
		if len(get_settings().Users) == 0 {
			return c.SendStatus(fiber.StatusNotFound)
		} else {
			return c.SendStatus(fiber.StatusOK)
		}
	})
	app.Get("/isUser-qm", func(c *fiber.Ctx) error {
		log(ATTEMPT, "CLIENT: Are these dudes with you?", true)
		if check_auth(c) {
			return c.SendStatus(fiber.StatusOK)
		} else {
			return c.SendStatus(fiber.StatusForbidden)
		}
	})
	app.Static("/logo", "./logo.png")
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	// yrFiles
	if slices.Contains(get_settings().Services, "files") {
		var yrFilesSub, _ = fs.Sub(yrFiles, "gui/yrFiles")
		app.Get("/yrFiles/", func(c *fiber.Ctx) error {
			if !check_allowed(c) {
				return c.SendStatus(fiber.StatusForbidden)
			}
			return c.Next()
		})
		app.Get("/yrFiles/files/+", http_serveFile)
		app.Use("/yrFiles/", filesystem.New(filesystem.Config{Root: http.FS(yrFilesSub)}))
		app.Get("/yrFiles", func(c *fiber.Ctx) error {
			return c.Redirect("/yrFiles/", fiber.StatusMovedPermanently)
		})
		app.Get("/yrText/", func(c *fiber.Ctx) error {
			if !check_allowed(c) || !check_auth(c) {
				return c.SendStatus(fiber.StatusForbidden)
			}
			return c.Next()
		})
		var yrTextSub, _ = fs.Sub(yrText, "gui/yrText")
		app.Use("/yrText/", filesystem.New(filesystem.Config{Root: http.FS(yrTextSub)}))

		app.Get("/delete-file", http_deleteFile)
		app.Get("/get-files/*", http_getFiles)
		app.Get("/rename-file", http_renameFile)
		app.Post("/upload-chunk", http_uploadChunk)
		app.Post("/write-to-file", http_writeToFile)
	}
	// yrSound
	if slices.Contains(get_settings().Services, "sound") {
		var yrSoundSub, _ = fs.Sub(yrSound, "gui/yrSound")
		app.Use("/yrSound/", func(c *fiber.Ctx) error {
			if !check_allowed(c) {
				return c.SendStatus(fiber.StatusForbidden)
			}
			return c.Next()
		})
		app.Use("/yrSound/", filesystem.New(filesystem.Config{Root: http.FS(yrSoundSub)}))
		app.Get("/yrSound", func(c *fiber.Ctx) error {
			return c.Redirect("/yrSound/", fiber.StatusMovedPermanently)
		})

		app.Get("/get-cover", http_getCover)
		app.Post("/get-id", http_getID)
		app.Get("/get-albums", http_getAlbums)
		app.Post("/get-album", http_getAlbum)
		app.Get("/get-songs", http_getSongs)
		app.Get("/get-artists", http_getArtists)
		app.Get("/get-artist-picture", http_getArtistPFP)
		app.Get("/get-song-info", http_getSongInfo)
		app.Get("/get-song-blob", http_getSongBlob)
	}

	// log for start
	fmt.Print("\033c")
	fmt.Printf(GREEN+"Server initiated!"+RESET+" Serving at "+PINK+"%s"+RESET+"\n", listener.Addr())
	fmt.Println("Press " + PINK + "<Ctrl-C>" + RESET + " to stop the server.")
	fmt.Println(PINK + "-------------------------------------------------" + RESET)

	app.Listener(listener)
}

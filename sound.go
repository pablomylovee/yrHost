package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
	"github.com/gofiber/fiber/v2"
)

type Song struct {
	FilePath    string `json:"file-path"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	AlbumArtist string `json:"album-artist"`
	Year        int    `json:"year"`
	Track       int    `json:"track-number"`
	Disc        int    `json:"disc-number"`
}
type SongwID struct {
	ID          int    `json:"id"`
	FilePath    string `json:"file-path"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	AlbumArtist string `json:"album-artist"`
	Year        int    `json:"year"`
	Track       int    `json:"track-number"`
	Disc        int    `json:"disc-number"`
}
type Album struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Year   int    `json:"year"`
}

var ys_savePath string = get_settings().YrSound.SavePath

func Index(path string, songsSlice *[]Song) error {
	var stats, err1 = os.Stat(path)
	if err1 != nil {
		return err1
	}

	if stats.Mode().IsDir() {
		var contents, err4 = os.ReadDir(path)
		if err4 != nil {
			return err4
		}
		var toIndex []os.DirEntry

		for _, v := range contents {
			if v.IsDir() {
				toIndex = append(toIndex, v)
			} else if strings.HasSuffix(v.Name(), ".mp3") || strings.HasSuffix(v.Name(), ".mp4") {
				var file, err2 = os.Open(filepath.Join(path, v.Name()))
				if err2 != nil {
					return err2
				}
				var songRaw, err3 = tag.ReadFrom(file)
				if err3 != nil {
					return err3
				}
				var discn, _ = songRaw.Disc()
				var trackn, _ = songRaw.Track()
				var title, artist, album, albumArtist string
				if songRaw.Title() != "" {
					title = songRaw.Title()
				} else {
					title = v.Name()
				}
				if songRaw.Artist() != "" {
					artist = songRaw.Artist()
				} else {
					artist = "Unknown artist"
				}
				if songRaw.Album() != "" {
					album = songRaw.Album()
				} else {
					album = "Unknown album"
				}
				if songRaw.AlbumArtist() != "" {
					albumArtist = songRaw.AlbumArtist()
				} else {
					albumArtist = "Unknown artist"
				}
				*songsSlice = append(*songsSlice, Song{
					FilePath:    filepath.Join(path, v.Name()),
					Title:       title,
					Artist:      artist,
					Album:       album,
					AlbumArtist: albumArtist,
					Year:        songRaw.Year(),
					Disc:        discn,
					Track:       trackn,
				})
			}
		}

		for _, v := range toIndex {
			Index(filepath.Join(path, v.Name()), songsSlice)
		}
	}
	return nil
}

func http_getCover(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}

	log(ATTEMPT, "Attempt to get song cover initiated.", false)
	var username string = c.Query("username")
	var id, _ = strconv.Atoi(c.Query("id"))
	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, username, ".db"))
	if err != nil {
		log(ERROR, "An error occured while trying to open yrSound database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()
	var song, err1 = ysDB.Query(`select title, artist, filepath from songs where id=?`, id)
	if err1 != nil {
		log(ERROR, "An error occured while trying to find song.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer song.Close()
	if song.Next() {
		var title, artist, file_path string
		song.Scan(&title, &artist, &file_path)
		log(COMPLETE, fmt.Sprintf("Found song! Found: %s - %s", title, artist), false)

		var file, err3 = os.Open(file_path)
		if err3 != nil {
			log(ERROR, "A system error occured while trying to open song file.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		defer file.Close()
		var songMD, err2 = tag.ReadFrom(file)
		if err2 != nil {
			log(ERROR, "A system error occured while tring to open song file.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if songMD.Picture() != nil {
			var mimetype string = http.DetectContentType(songMD.Picture().Data[:600])
			c.Type(mimetype)
			log(COMPLETE, "Returned song cover successfully!", true)
			return c.Send(songMD.Picture().Data)
		} else {
			log(ERROR, "Song didn't have a cover.", true)
			return c.SendStatus(fiber.StatusNotFound)
		}

	} else {
		log(ERROR, "Couldn't find song.", true)
		return c.SendStatus(fiber.StatusNotFound)
	}
}

func http_getID(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to get song ID initiated.", false)

	var body []byte = c.Body()

	var username string = c.Query("username")
	var songInfo struct {
		Title       string `json:"title"`
		Artist      string `json:"artist"`
		Album       string `json:"album"`
		AlbumArtist string `json:"album-artist"`
	}
	json.Unmarshal(body, &songInfo)

	var ysDB, err1 = sql.Open("sqlite3", filepath.Join(ys_savePath, username, ".db"))
	if err1 != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()

	var song, err2 = ysDB.Query(`select id from songs where title=? AND artist=? AND album=? AND albumArtist=?`,
		songInfo.Title, songInfo.Artist, songInfo.Album, songInfo.AlbumArtist,
	)
	if err2 != nil {
		log(ERROR, "A system error occured while trying to search for song ID.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer song.Close()

	if song.Next() {
		var id int64
		var err error = song.Scan(&id)
		if err != nil {
			log(ERROR, "An error occured while trying to get song ID from database.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		var idStr string = strconv.FormatInt(id, 10)
		c.Type("text/plain")
		log(COMPLETE, "Song ID returned!", true)
		return c.Send([]byte(idStr))
	} else {
		log(ERROR, "Could not find song.", true)
		return c.SendStatus(fiber.StatusNotFound)
	}
}

func http_getAlbums(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to get albums initiated.", false)

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, c.Query("username"), ".db"))
	if err != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()

	var albums []Album
	var query, err1 = ysDB.Query(`select album, albumArtist, year from songs`)
	if err1 != nil {
		log(ERROR, "A system error occured while trying to get albums from database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer query.Close()

	for query.Next() {
		var year int
		var album, albumArtist string
		var err3 error = query.Scan(&album, &albumArtist, &year)
		if err3 != nil {
			log(ERROR, "An error occured while getting albums.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		var albumType Album = Album{
			Title: album, Artist: albumArtist,
			Year: year,
		}
		var exists bool = false
		for _, v := range albums {
			if albumType.Title == v.Title &&
				albumType.Artist == v.Artist &&
				albumType.Year == v.Year {
				exists = true
			}
		}
		if !exists {
			albums = append(albums, albumType)
		}
	}

	c.Type("application/json")
	var content, err2 = json.Marshal(albums)
	if err2 != nil {
		log(ERROR, "An error occured while trying to write albums to user.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	log(COMPLETE, "Returned all albums!", true)
	return c.Send(content)
}

func http_getAlbum(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to get album info initiated.", false)

	var body []byte = c.Body()
	var args struct {
		Album       string `json:"album"`
		AlbumArtist string `json:"album-artist"`
	}
	json.Unmarshal(body, &args)

	var ysDB, err1 = sql.Open("sqlite3", filepath.Join(ys_savePath, c.Query("username"), ".db"))
	if err1 != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()

	var album, err2 = ysDB.Query(`select * from songs where album=? AND albumArtist=?`, args.Album, args.AlbumArtist)
	if err2 != nil {
		log(ERROR, "An error occured while trying to get album's songs", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer album.Close()

	var songs []SongwID
	for album.Next() {
		var id, year, track, disc int
		var file_path, title, artist, Album, albumArtist string

		album.Scan(&id, &file_path, &title, &artist, &Album, &albumArtist, &year, &track, &disc)
		songs = append(songs, SongwID{
			ID:          id,
			FilePath:    file_path,
			Title:       title,
			Artist:      artist,
			Album:       Album,
			AlbumArtist: albumArtist,
			Year:        year,
			Track:       track,
			Disc:        disc,
		})
	}

	var songsByte, err3 = json.Marshal(songs)
	if err3 != nil {
		log(ERROR, "An error occured while trying to encode songs to JSON.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	c.Type("application/json")
	log(COMPLETE, "Successfully returned songs from the album!", true)
	return c.Send(songsByte)
}

func http_getSongs(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to get all songs initiated", false)

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, c.Query("username"), ".db"))
	if err != nil {
		log(ERROR, "An error occured while trying to open yrSound database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()

	var rows, err1 = ysDB.Query(`select * from songs`)
	if err1 != nil {
		log(ERROR, "An error occured while trying to get all songs.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer rows.Close()

	var songs []SongwID
	for rows.Next() {
		var id, year, track, disc int
		var file_path, title, artist, album, albumArtist string

		var err2 error = rows.Scan(&id, &file_path, &title, &artist, &album, &albumArtist, &year, &track, &disc)
		if err2 != nil {
			log(ERROR, "An error occured while trying to get song values.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		var song SongwID = SongwID{
			ID:          id,
			FilePath:    file_path,
			Title:       title,
			Artist:      artist,
			Album:       album,
			AlbumArtist: albumArtist,
			Year:        year,
			Track:       track,
			Disc:        disc,
		}
		songs = append(songs, song)
	}

	c.Type("application/json")
	var songsByte, err3 = json.Marshal(songs)
	if err3 != nil {
		log(ERROR, "An error occured while trying to encode songs to JSON.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	log(COMPLETE, "Returned all songs successfully!", true)
	return c.Send(songsByte)
}

func http_getArtists(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to all artists initiated.", false)

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, c.Query("username"), ".db"))
	if err != nil {
		log(ERROR, "An error occured while trying to open yrSound database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()

	var artists []string
	var rows, err2 = ysDB.Query(`select artist from songs`)
	if err2 != nil {
		log(ERROR, "An error occured while trying to get artists.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer rows.Close()
	for rows.Next() {
		var artist string
		rows.Scan(&artist)

		var exists bool = false
		for _, v := range artists {
			if artist == v {
				exists = true
			}
		}
		if !exists {
			artists = append(artists, artist)
		}
	}

	var artistsByte, err1 = json.Marshal(artists)
	if err1 != nil {
		log(ERROR, "An error occured while trying to encode artists to JSON.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	c.Type("application/json")
	log(COMPLETE, "All artists successfully returned!", true)
	return c.Send(artistsByte)
}

func http_getSongInfo(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to get song info initiated.", false)
	var id, _ = strconv.Atoi(c.Query("id"))

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, c.Query("username"), ".db"))
	if err != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()

	var song, err1 = ysDB.Query(`select * from songs where id=?`, id)
	if err1 != nil {
		log(ERROR, "An error occured while trying to search for song.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()

	if song.Next() {
		var id, year, track, disc int
		var file_path, title, artist, album, albumArtist string
		var err3 error = song.Scan(&id, &file_path, &title, &artist, &album, &albumArtist, &year, &track, &disc)
		if err3 != nil {
			log(ERROR, "An error occured while trying to get song info.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		var songByte, err2 = json.Marshal(Song{
			FilePath:    file_path,
			Title:       title,
			Artist:      artist,
			Album:       album,
			AlbumArtist: albumArtist,
			Year:        year,
			Track:       track,
			Disc:        disc,
		})
		if err2 != nil {
			log(ERROR, "An error occured while trying to convert song to JSON format", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		c.Type("application/json")
		log(COMPLETE, "Returned song info successfully.", true)
		return c.Send(songByte)
	} else {
		log(ERROR, "Could not find song.", true)
		return c.SendStatus(fiber.StatusNotFound)
	}
}

func http_getSongBlob(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to get song blob initiated.", false)

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, c.Query("username"), ".db"))
	if err != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer ysDB.Close()

	var id, _ = strconv.Atoi(c.Query("id"))
	var song, err1 = ysDB.Query(`select filepath from songs where id=?`, id)
	if err1 != nil {
		log(ERROR, "An error occured while trying to get song's file path.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer song.Close()

	if song.Next() {
		var file_path string
		var err error = song.Scan(&file_path)
		if err != nil {
			log(ERROR, "An error occured while trying to get song's file path.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		var songFile, err1 = os.Open(file_path)
		if err1 != nil {
			log(ERROR, "An error occured while trying to open song.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		defer songFile.Close()

		var songBytes, err2 = io.ReadAll(songFile)
		if err2 != nil {
			log(ERROR, "An error occured while trying to read song.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		c.Type("application/json")
		log(COMPLETE, "Returned song blob successfully!", true)
		return c.Send(songBytes)
	} else {
		log(ERROR, "Couldn't find song.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
}

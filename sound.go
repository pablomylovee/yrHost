package main

import (
	"bytes"
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

func http_getCover(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}

	log(ATTEMPT, "Attempt to get song cover initiated.", false)
	var username string = r.URL.Query().Get("username")
	var id, _ = strconv.Atoi(r.URL.Query().Get("id"))
	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, username, ".db"))
	if err != nil {
		log(ERROR, "An error occured while trying to open yrSound database.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()
	var song, err1 = ysDB.Query(`select title, artist, filepath from songs where id=?`, id)
	if err1 != nil {
		log(ERROR, "An error occured while trying to find song.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer song.Close()
	if song.Next() {
		var title, artist, file_path string
		song.Scan(&title, &artist, &file_path)
		log(COMPLETE, fmt.Sprintf("Found song! Found: %s - %s", title, artist), false)

		var file, err3 = os.Open(file_path)
		if err3 != nil {
			log(ERROR, "A system error occured while trying to open song file.", true)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()
		var songMD, err2 = tag.ReadFrom(file)
		if err2 != nil {
			log(ERROR, "A system error occured while tring to open song file.", true)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if songMD.Picture() != nil {
			var mimetype string = http.DetectContentType(songMD.Picture().Data[:600])
			w.Header().Set("Content-Type", mimetype)
			var picbuffer *bytes.Buffer = bytes.NewBuffer(songMD.Picture().Data)
			io.CopyBuffer(w, picbuffer, make([]byte, 3072))
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log(COMPLETE, "Returned song cover successfully!", true)
	} else {
		log(ERROR, "Couldn't find song.", true)
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func http_getID(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to get song ID initiated.", false)

	defer r.Body.Close()
	var body, err = io.ReadAll(r.Body)
	if err != nil {
		log(ERROR, "An error occured while trying to read request body.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var username string = r.URL.Query().Get("username")
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()

	var song, err2 = ysDB.Query(`select id from songs where title=? AND artist=? AND album=? AND albumArtist=?`,
		songInfo.Title, songInfo.Artist, songInfo.Album, songInfo.AlbumArtist,
	)
	if err2 != nil {
		log(ERROR, "A system error occured while trying to search for song ID.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer song.Close()

	if song.Next() {
		var id int64
		var err error = song.Scan(&id)
		if err != nil {
			return
		}
		var idStr string = strconv.FormatInt(id, 10)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(idStr))

		log(COMPLETE, "Song ID returned!", true)
	} else {
		log(ERROR, "Could not find song.", true)
		w.WriteHeader(http.StatusNotFound)
	}
}

func http_getAlbums(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to get albums initiated.", false)

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, r.URL.Query().Get("username"), ".db"))
	if err != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()

	var albums []Album
	var query, err1 = ysDB.Query(`select album, albumArtist, year from songs`)
	if err1 != nil {
		log(ERROR, "A system error occured while trying to get albums from database.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer query.Close()

	for query.Next() {
		var year int
		var album, albumArtist string
		var err3 error = query.Scan(&album, &albumArtist, &year)
		if err3 != nil {
			log(ERROR, "An error occured while getting albums.", true)
			w.WriteHeader(http.StatusInternalServerError)
			return
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

	w.Header().Set("Content-Type", "application/json")
	var content, err2 = json.Marshal(albums)
	if err2 != nil {
		log(ERROR, "An error occured while trying to write albums to user.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var contentBuffer *bytes.Buffer = bytes.NewBuffer(content)
	io.CopyBuffer(w, contentBuffer, make([]byte, 3072))
	log(COMPLETE, "Returned all albums!", true)
}

func http_getAlbum(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to get album info initiated.", false)

	defer r.Body.Close()
	var body, err = io.ReadAll(r.Body)
	if err != nil {
		log(ERROR, "An error occured while trying to read request body.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var args struct {
		Album       string `json:"album"`
		AlbumArtist string `json:"album-artist"`
	}
	json.Unmarshal(body, &args)

	var ysDB, err1 = sql.Open("sqlite3", filepath.Join(ys_savePath, r.URL.Query().Get("username"), ".db"))
	if err1 != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()

	var album, err2 = ysDB.Query(`select * from songs where album=? AND albumArtist=?`, args.Album, args.AlbumArtist)
	if err2 != nil {
		log(ERROR, "An error occured while trying to get album's songs", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var songsBuffer *bytes.Buffer = bytes.NewBuffer(songsByte)
	w.Header().Set("Content-Type", "application/json")
	io.CopyBuffer(w, songsBuffer, make([]byte, 3072))
	log(COMPLETE, "Successfully returned songs from the album!", true)
}

func http_getSongs(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to get all songs initiated", false)

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, r.URL.Query().Get("username"), ".db"))
	if err != nil {
		log(ERROR, "An error occured while trying to open yrSound database.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()

	var rows, err1 = ysDB.Query(`select * from songs`)
	if err1 != nil {
		log(ERROR, "An error occured while trying to get all songs.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var songs []SongwID
	for rows.Next() {
		var id, year, track, disc int
		var file_path, title, artist, album, albumArtist string

		var err2 error = rows.Scan(&id, &file_path, &title, &artist, &album, &albumArtist, &year, &track, &disc)
		if err2 != nil {
			log(ERROR, "An error occured while trying to get song values.", true)
			w.WriteHeader(http.StatusInternalServerError)
			return
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
		fmt.Println(song)
		songs = append(songs, song)
	}
	fmt.Println(songs)

	w.Header().Set("Content-Type", "application/json")
	var songsByte, err3 = json.Marshal(songs)
	if err3 != nil {
		log(ERROR, "An error occured while trying to encode songs to JSON.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var songsBuffer *bytes.Buffer = bytes.NewBuffer(songsByte)
	io.CopyBuffer(w, songsBuffer, make([]byte, 3072))

	log(COMPLETE, "Returned all songs successfully!", true)
}

func http_getArtists(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to all artists initiated.", false)

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, r.URL.Query().Get("username"), ".db"))
	if err != nil {
		log(ERROR, "An error occured while trying to open yrSound database.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()

	var artists []string
	var rows, err2 = ysDB.Query(`select artist from songs`)
	if err2 != nil {
		log(ERROR, "An error occured while trying to get artists.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var artistsBuffer *bytes.Buffer = bytes.NewBuffer(artistsByte)
	w.Header().Set("Content-Type", "application/json")
	io.CopyBuffer(w, artistsBuffer, make([]byte, 3072))

	log(COMPLETE, "All artists successfully returned!", true)
}

func http_getSongInfo(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to get song info initiated.", false)
	var id, _ = strconv.Atoi(r.URL.Query().Get("id"))

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, r.URL.Query().Get("username"), ".db"))
	if err != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()

	var song, err1 = ysDB.Query(`select * from songs where id=?`, id)
	if err1 != nil {
		log(ERROR, "An error occured while trying to search for song.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()

	if song.Next() {
		var id, year, track, disc int
		var file_path, title, artist, album, albumArtist string
		var err3 error = song.Scan(&id, &file_path, &title, &artist, &album, &albumArtist, &year, &track, &disc)
		if err3 != nil {
			log(ERROR, "An error occured while trying to get song info.", true)
			w.WriteHeader(http.StatusInternalServerError)
			return
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var songBuffer *bytes.Buffer = bytes.NewBuffer(songByte)
		w.Header().Set("Content-Type", "application/json")
		io.CopyBuffer(w, songBuffer, make([]byte, 3072))
		log(COMPLETE, "Returned song info successfully.", true)
	} else {
		log(ERROR, "Could not find song.", true)
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func http_getSongBlob(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to get song blob initiated.", false)

	var ysDB, err = sql.Open("sqlite3", filepath.Join(ys_savePath, r.URL.Query().Get("username"), ".db"))
	if err != nil {
		log(ERROR, "A system error occured while trying to open yrSound database.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ysDB.Close()

	var id, _ = strconv.Atoi(r.URL.Query().Get("id"))
	var song, err1 = ysDB.Query(`select filepath from songs where id=?`, id)
	if err1 != nil {
		log(ERROR, "An error occured while trying to get song's file path.", true)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer song.Close()

	if song.Next() {
		var file_path string
		var err error = song.Scan(&file_path)
		if err != nil {
			log(ERROR, "An error occured while trying to get song's file path.", true)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var songFile, err1 = os.Open(file_path)
		if err1 != nil {
			log(ERROR, "An error occured while trying to open song.", true)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer songFile.Close()

		var songBytes, err2 = io.ReadAll(songFile)
		if err2 != nil {
			log(ERROR, "An error occured while trying to read song.", true)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		songFile.Seek(0, io.SeekStart)
		w.Header().Set("Content-Type", http.DetectContentType(songBytes[:600]))
		io.CopyBuffer(w, songFile, make([]byte, 3072))

		log(COMPLETE, "Returned song blob successfully!", true)
	} else {
		log(ERROR, "Couldn't find song.", true)
		w.WriteHeader(http.StatusNotFound)
	}
}

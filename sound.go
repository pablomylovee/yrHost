package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
type Album struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Year   int    `json:"year"`
	Songs  []Song `json:"songs"`
}

var ys_savePath string = get_settings().YrSound.SavePath

func getPath(toSearch Song, songsSlice []Song) (string, bool) {
	for _, song := range songsSlice {
		if toSearch.Title == song.Title &&
			toSearch.Artist == song.Artist &&
			toSearch.Album == song.Album &&
			toSearch.AlbumArtist == song.AlbumArtist {
			return song.FilePath, true
		}
	}
	return "", false
}
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
				*songsSlice = append(*songsSlice, Song{
					FilePath:    filepath.Join(path, v.Name()),
					Title:       songRaw.Title(),
					Artist:      songRaw.Artist(),
					Album:       songRaw.Album(),
					AlbumArtist: songRaw.AlbumArtist(),
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
	var songs []Song
	var err2 error = Index(filepath.Join(ys_savePath, r.URL.Query().Get("username")), &songs)
	if err2 != nil {
		fmt.Println(err2.Error())
		return
	}

	defer r.Body.Close()
	var body, _ = io.ReadAll(r.Body)
	var song Song
	json.Unmarshal(body, &song)

	var songPath, found = getPath(song, songs)
	if !found {
		log(ERROR, "Song couldn't be found.", true)
		return
	}
	var songFile, _ = os.Open(songPath)
	log(COMPLETE, "Song found!", false)

	log(ATTEMPT, "Getting song information.", false)
	var songMD, err = tag.ReadFrom(songFile)
	if err != nil {
		log(ERROR, "A system error occured while trying to get song information.", true)
		return
	}
	log(COMPLETE, "Song information received!", false)

	log(ATTEMPT, "Returning song cover.", false)
	var cover io.Reader = bytes.NewBuffer(songMD.Picture().Data)
	w.Header().Set("Content-Type", songMD.Picture().MIMEType)
	fmt.Println(songMD.Picture().MIMEType)
	var _, err1 = io.CopyBuffer(w, cover, make([]byte, 3072))
	if err1 != nil {
		log(ERROR, "An error occured while trying to return song cover.", true)
		return
	}
	log(COMPLETE, "Song cover successfully returned!", true)
}

func http_getAlbums(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to get albums initiated.", false)

	var songs []Song
	var err error = Index(filepath.Join(ys_savePath, r.URL.Query().Get("username")), &songs)
	if err != nil {
		log(ERROR, "An error occured while indexing songs.", true)
		return
	}
	log(COMPLETE, "Indexed songs!", false)

	log(ATTEMPT, "Sorting songs to their albums...", false)
	var albums []Album
	for _, song := range songs {
		var added bool = false
		for i := range albums {
			if song.Album == albums[i].Title {
				albums[i].Songs = append(albums[i].Songs, song)
				added = true
				break
			}
		}
		if !added {
			albums = append(albums, Album{
				Title:  song.Album,
				Artist: song.AlbumArtist,
				Year:   song.Year,
				Songs:  []Song{song},
			})
		}
	}
	log(COMPLETE, "Songs sorted successfully!", false)

	log(ATTEMPT, "Returning JSON-formatted result.", false)
	var albumsBytes, _ = json.Marshal(albums)
	var albumsBuffer *bytes.Buffer = bytes.NewBuffer(albumsBytes)

	w.Header().Set("Content-Type", "application/json")
	io.CopyBuffer(w, albumsBuffer, make([]byte, 3072))
	log(COMPLETE, "Albums returned successfully!", true)
}

func http_getSongs(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	log(ATTEMPT, "Attempt to get all songs initiated.", false)

	var songs []Song
	var err error = Index(filepath.Join(ys_savePath, r.URL.Query().Get("username")), &songs)
	if err != nil {
		log(ERROR, "An error occured while trying to get song names.", true)
	}

	log(COMPLETE, "Got all songs!", false)
	log(ATTEMPT, "Returning songs list.", false)

	var songsBytes, _ = json.Marshal(songs)
	var songsBuffer *bytes.Buffer = bytes.NewBuffer(songsBytes)

	w.Header().Set("Content-Type", "application/json")
	io.CopyBuffer(w, songsBuffer, make([]byte, 3072))
	log(COMPLETE, "Songs returned successfully!", true)
}

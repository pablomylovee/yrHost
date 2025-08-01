package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type DirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var yf_savePath string = get_settings().YrFiles.SavePath

func http_deleteFile(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	var file_td string = r.URL.Query().Get("name")
	var username string = r.URL.Query().Get("username")
	log(ATTEMPT, "Attempt to delete a file initiated. ("+get_datentime()+")", false)

	if file_td == "" {
		log(ERROR, "No target returned", true)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var parts []string = strings.Split(file_td, "~/~")
	file_td = filepath.Clean(filepath.Join(parts...))
	var ts_filePath string = filepath.Join(yf_savePath, username, file_td)
	log(COMPLETE, "Path created: "+ts_filePath, false)

	var _, err = os.Stat(ts_filePath)
	if errors.Is(err, fs.ErrNotExist) {
		log(ERROR, "Nonexistent file returned", true)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := os.RemoveAll(ts_filePath); err != nil {
		log(ERROR, "A system error occured while attempting to remove '"+ts_filePath+"'\n	error: "+err.Error()+".", true)
		w.WriteHeader(http.StatusForbidden)
		return
	} else {
		log(COMPLETE, "File removal was successful.", true)
		w.WriteHeader(http.StatusFound)
		return
	}
}

func http_renameFile(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	var username string = r.URL.Query().Get("username")
	var path string = r.URL.Query().Get("path")
	var new_name string = r.URL.Query().Get("newname")
	log(ATTEMPT, "Attempt to rename a file initiated.", false)

	if path == "" || new_name == "" || (path == "" && new_name == "") {
		log(ERROR, "No target file returned.", true)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var parts []string = strings.Split(path, "~/~")
	path = filepath.Join(parts...)
	var target_file string = filepath.Join(yf_savePath, username, path)
	log(COMPLETE, "Path created: "+target_file, false)

	var dir_name, _ = filepath.Split(target_file)
	var to_path string = filepath.Join(dir_name, new_name)
	log(COMPLETE, "New path created: "+target_file, false)
	var _, err1 = os.Stat(to_path)
	if errors.Is(err1, fs.ErrNotExist) {
		var err error = os.Rename(target_file, to_path)

		if err != nil {
			log(ERROR, "A system error occured while trying to rename file.", true)
			w.WriteHeader(http.StatusForbidden)
			return
		} else {
			log(COMPLETE, "File renaming was successful. ("+get_datentime()+")", true)
			w.WriteHeader(http.StatusOK)
			return
		}
	} else {
		log(ERROR, "A taken name was returned.", false)
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func http_uploadChunk(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	var username string = r.URL.Query().Get("username")
	var filename string = r.URL.Query().Get("filename")
	log(ATTEMPT, "Attempt to write chunks initiated.", false)

	r.Body = http.MaxBytesReader(w, r.Body, 10<<30)
	var target_path string = filepath.Join(yf_savePath, username, filepath.Join(strings.Split(filename, "~/~")...))
	os.MkdirAll(filepath.Dir(target_path), 0755)

	var dest, err1 = os.OpenFile(target_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err1 != nil {
		log(ERROR, "A system error occured while trying to open target file.", true)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	defer dest.Close()
	var _, err2 = io.Copy(dest, r.Body)
	if err2 != nil {
		log(ERROR, "A system error occured while trying to write to target file.", true)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	log(COMPLETE, "Successfully wrote chunks!", true)
}

func http_getContent(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	var username string = r.URL.Query().Get("username")
	var path string = r.URL.Query().Get("path")

	var parts []string = strings.Split(path, "~/~")
	path = filepath.Join(parts...)
	log(ATTEMPT, "Working on: "+path, false)
	var fullpath string = filepath.Join(yf_savePath, username, path)
	log(COMPLETE, "Target path created: "+fullpath, false)

	var file, err = os.Open(fullpath)
	if err != nil {
		log(ERROR, "A system error occured while trying to read target file.", true)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	var fileContent, _ = os.ReadFile(fullpath)
	log(COMPLETE, "Successfully read file.", false)
	var contentType string = http.DetectContentType(fileContent[:512])
	log(ATTEMPT, "File type received: "+contentType, false)
	w.Header().Set("Content-Type", contentType)
	io.CopyBuffer(w, file, make([]byte, 3072))
	log(COMPLETE, "File content returned! ("+get_datentime()+")", true)
}

func http_getFiles(w http.ResponseWriter, r *http.Request) {
	if !check_allowed(r) || !check_auth(r) {
		return
	}
	var username string = r.URL.Query().Get("username")
	var search_dir string = r.URL.Query().Get("dir")
	log(ATTEMPT, "Attempt to get files from a directory.", false)

	log(ATTEMPT, "Directory returned: '"+search_dir+"'", false)
	var parts []string = strings.Split(search_dir, "~/~")
	search_dir = filepath.Clean(filepath.Join(parts...))
	log(COMPLETE, "Relative path created: '"+search_dir+"'", false)

	var toReturn []DirEntry
	var files []os.DirEntry

	if search_dir == "" {
		var search_path string = filepath.Join(yf_savePath, username)
		log(COMPLETE, "Full path created: "+search_path, false)
		var x, err = os.ReadDir(search_path)
		if err != nil {
			log(ERROR, "A system error occured during directory indexing.", true)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		files = x
	} else {
		var search_path string = filepath.Join(yf_savePath, username, search_dir)
		log(COMPLETE, "Full path created: "+search_path, false)
		var x, err = os.ReadDir(search_path)
		if err != nil {
			log(ERROR, "A system error occured during directory indexing.", true)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		files = x
	}

	for _, v := range files {
		var file DirEntry
		file.Name = v.Name()
		if v.Type().IsDir() {
			file.Type = "d"
		} else {
			file.Type = "f"
		}

		toReturn = append(toReturn, file)
	}

	log(COMPLETE, "All files stored in a list.", false)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toReturn)
	log(COMPLETE, "Directory indexing was successful ("+get_datentime()+")", true)
}

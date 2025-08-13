package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
)

type DirEntry struct {
	Name         string `json:"name"`
	RelativePath string `json:"relative-path"`
	Type         string `json:"type"`
}

var yf_savePath string = get_settings().YrFiles.SavePath

func http_deleteFile(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	var file_td string = c.Query("name")
	var username string = c.Query("username")
	log(ATTEMPT, "Attempt to delete a file initiated. ("+get_datentime()+")", false)

	if file_td == "" {
		log(ERROR, "No target returned", true)
		return c.SendStatus(fiber.StatusNotFound)
	}
	var parts []string = strings.Split(file_td, "~/~")
	file_td = filepath.Clean(filepath.Join(parts...))
	var ts_filePath string = filepath.Join(yf_savePath, username, file_td)
	log(COMPLETE, "Path created: "+ts_filePath, false)

	var _, err = os.Stat(ts_filePath)
	if errors.Is(err, fs.ErrNotExist) {
		log(ERROR, "Nonexistent file returned", true)
		return c.SendStatus(fiber.StatusNotFound)
	}

	if err := os.RemoveAll(ts_filePath); err != nil {
		log(ERROR, "A system error occured while attempting to remove '"+ts_filePath+"'\n	error: "+err.Error()+".", true)
		return c.SendStatus(fiber.StatusNotFound)
	} else {
		log(COMPLETE, "File removal was successful.", true)
		return c.SendStatus(fiber.StatusOK)
	}
}

func http_renameFile(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	var username string = c.Query("username")
	var path string = c.Query("path")
	var new_name string = c.Query("newname")
	log(ATTEMPT, "Attempt to rename a file initiated.", false)

	if path == "" || new_name == "" {
		log(ERROR, "No target file returned.", true)
		return c.SendStatus(fiber.StatusNotFound)
	}

	var parts []string = strings.Split(path, "~/~")
	path = filepath.Join(parts...)
	var target_file string = filepath.Join(yf_savePath, username, path)
	log(COMPLETE, "Path created: "+target_file, false)

	var dir_name, _ = filepath.Split(target_file)
	var to_path string = filepath.Join(dir_name, new_name)
	log(COMPLETE, "New path created: "+to_path, false)
	var stat, err1 = os.Stat(to_path)
	if errors.Is(err1, fs.ErrNotExist) {
		var err error = os.Rename(target_file, to_path)

		if err != nil {
			log(ERROR, "A system error occured while trying to rename file.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		} else {
			log(COMPLETE, "File renaming was successful. ("+get_datentime()+")", true)
			return c.SendStatus(fiber.StatusOK)
		}
	} else {
		if stat.IsDir() {
			log(ERROR, "A taken name was returned. (directory)", true)
		} else {
			log(ERROR, "A taken name was returned. (file)", true)
		}
		return c.SendStatus(fiber.StatusBadRequest)
	}
}

func http_uploadChunk(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	var username string = c.Query("username")
	var filename string = c.Query("filename")
	log(ATTEMPT, "Attempt to write chunks initiated.", false)

	var target_path string = filepath.Join(yf_savePath, username, filepath.Join(strings.Split(filename, "~/~")...))
	os.MkdirAll(filepath.Dir(target_path), 0755)

	var dest, err1 = os.OpenFile(target_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err1 != nil {
		log(ERROR, "A system error occured while trying to open target file.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer dest.Close()
	var body *bytes.Buffer = bytes.NewBuffer(c.Body())

	var _, err2 = io.Copy(dest, body)
	if err2 != nil {
		log(ERROR, "A system error occured while trying to write to target file.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	log(COMPLETE, "Successfully wrote chunks!", true)
	return c.SendStatus(fiber.StatusOK)
}
func http_writeToFile(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to write to file initiated.", false)

	var relativePath, _ = url.PathUnescape(c.Query("path"))
	var path string = filepath.Join(yf_savePath, c.Query("username"), relativePath)

	var dest, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log(ERROR, "A system error occured while trying to open target file.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer dest.Close()
	log(COMPLETE, "Opened file for writing!", false)

	log(ATTEMPT, "Creating and copying request body buffer to target.", false)
	log(STEP, fmt.Sprintf("Content-Length: %d", len(c.Body())), false)
	var buffer *bytes.Buffer = bytes.NewBuffer(c.Body())
	var w, err1 = io.CopyBuffer(dest, buffer, make([]byte, 3145728))
	if err1 != nil {
		log(ERROR, "A system error occured while trying to write to file.\n		"+err1.Error(), true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	log(STEP, fmt.Sprintf("Written: %d", w), false)

	log(COMPLETE, "Successfully wrote to file!", true)
	return c.SendStatus(fiber.StatusOK)
}

func http_serveFile(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	log(ATTEMPT, "Attempt to get file initiated.", false)
	var username string = c.Query("username")
	var rpath, _ = url.PathUnescape(c.Params("+"))
	var path string = filepath.Join(yf_savePath, username, rpath)

	var f, err = os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		log(ERROR, "Could not find requested file.", true)
		return c.SendStatus(fiber.StatusNotFound)
	}
	if f.IsDir() {
		log(ERROR, "Requested file is a directory. Use `getFiles`.", true)
		return c.SendStatus(fiber.StatusBadRequest)
	}

	var file, err1 = os.Open(path)
	if err1 != nil {
		log(ERROR, "A system error occured while trying to open requested file.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	defer file.Close()

	var buf []byte = make([]byte, 512)
	file.Read(buf)
	file.Seek(0, 0)

	var mimeType string = http.DetectContentType(buf)

	if mimeType == "application/octet-stream" {
		var ext string = strings.ToLower(filepath.Ext(filePath))
		if extType := mime.TypeByExtension(ext); extType != "" {
			mimeType = extType
		}
	}
	if mimeType == "application/octet-stream" {
		var mt, _ = mimetype.DetectReader(file)
		mimeType = mt.String()
	}
	c.Response().Header.Set("Content-Type", mimeType)
	log(COMPLETE, "Sent file successfully!", true)
	return c.SendFile(path)
}

func http_getFiles(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden)
	}
	var username string = c.Query("username")
	var search_dir, _ = url.PathUnescape(c.Params("*"))
	log(ATTEMPT, "Attempt to get files from a directory.", false)

	log(ATTEMPT, "Directory returned: '"+search_dir+"'", false)
	log(COMPLETE, "Relative path created: '"+search_dir+"'", false)

	var toReturn []DirEntry
	var files []os.DirEntry

	if search_dir == "" {
		var search_path string = filepath.Join(yf_savePath, username)
		log(COMPLETE, "Full path created: "+search_path, false)
		var x, err = os.ReadDir(search_path)
		if err != nil {
			log(ERROR, "A system error occured during directory indexing.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		files = x
	} else {
		var search_path string = filepath.Join(yf_savePath, username, search_dir)
		log(COMPLETE, "Full path created: "+search_path, false)
		var x, err = os.ReadDir(search_path)
		if err != nil {
			log(ERROR, "A system error occured during directory indexing.", true)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		files = x
	}

	for _, v := range files {
		var file DirEntry
		file.Name = v.Name()
		var parts []string = strings.Split(search_dir, "/")
		parts = slices.DeleteFunc(parts, func(s string) bool { return s == "" })
		file.RelativePath = filepath.Join(filepath.Join(parts...), v.Name())
		if v.Type().IsDir() {
			file.Type = "d"
		} else {
			file.Type = "f"
		}

		toReturn = append(toReturn, file)
	}

	log(COMPLETE, "All files stored in a list.", false)

	c.Type("application/json")
	var json, err = json.Marshal(toReturn)
	if err != nil {
		log(ERROR, "An error occured while trying to return files.", true)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	log(COMPLETE, "All files sent successfully!", true)
	return c.Send(json)
}

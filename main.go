package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type DirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type UserPreferences struct {
	Port      int      `json:"port"`
	UseAuth   bool     `json:"use-auth"`
	Auth      string   `json:"auth"`
	Blacklist []string `json:"ip-blacklist"`
}

var _, filePath, _, _ = runtime.Caller(0)
var (
	RED   string = "\033[31m"
	BLUE  string = "\033[34m"
	CYAN  string = "\033[36m"
	PINK  string = "\033[35m"
	GREEN string = "\033[32m"
	RESET string = "\033[0m"
)

// settings

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
func http_main(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	http.FileServer(http.Dir(filepath.Join(filePath, "web"))).ServeHTTP(w, r)
}
func http_explore(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	http.StripPrefix("/explorer", http.FileServer(http.Dir(filepath.Join(filePath, "web", "explorer")))).ServeHTTP(w, r)
}
func http_deleteFile(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if get_settings().UseAuth {
		var auth string = r.URL.Query().Get("auth")
		if !(auth == get_settings().Auth) {
			fmt.Println(RED + ">> " + RESET + "Invalid authentication.")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
	}
	var file_td string = r.URL.Query().Get("name")
	fmt.Println(CYAN + ">> " + RESET + "Attempt to delete a file initiated. (" + get_datentime() + ")")

	if file_td == "" {
		fmt.Println(RED + ">> " + RESET + "No target returned.")
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(PINK + "-------------------------------------------------" + RESET)
		return
	}
	var parts []string = strings.Split(file_td, "~/~")
	file_td = filepath.Clean(filepath.Join(parts...))
	var ts_filePath string = filepath.Join(filePath, "storage", file_td)
	fmt.Println(GREEN + ">> " + RESET + "Path created: " + ts_filePath)

	var _, err = os.Stat(ts_filePath)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println(RED + ">> " + RESET + "Nonexistent file returned.")
		w.WriteHeader(http.StatusNotFound)
		fmt.Println(PINK + "-------------------------------------------------" + RESET)
		return
	}

	if err := os.RemoveAll(ts_filePath); err != nil {
		fmt.Println(RED + ">> " + RESET + err.Error())
		w.WriteHeader(http.StatusForbidden)
		fmt.Println(PINK + "-------------------------------------------------" + RESET)
		return
	} else {
		fmt.Println(GREEN + ">> " + RESET + "File removal successful. (" + get_datentime() + ")")
		w.WriteHeader(http.StatusFound)
		fmt.Println(PINK + "-------------------------------------------------" + RESET)
		return
	}
}

func http_renameFile(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if get_settings().UseAuth {
		var auth string = r.URL.Query().Get("auth")
		if !(auth == get_settings().Auth) {
			fmt.Println(RED + ">> " + RESET + "Invalid authentication.")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
	}
	var path string = r.URL.Query().Get("path")
	var new_name string = r.URL.Query().Get("newname")
	fmt.Println(CYAN + ">> " + RESET + "Attempt to rename a file initiated. (" + get_datentime() + ")")

	if path == "" || new_name == "" || (path == "" && new_name == "") {
		fmt.Println(RED + ">> " + RESET + "No target returned.")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(PINK + "-------------------------------------------------" + RESET)
		return
	}

	var parts []string = strings.Split(path, "~/~")
	path = filepath.Join(parts...)
	var target_file string = filepath.Join(filePath, "storage", path)
	fmt.Println(GREEN + ">> " + RESET + "Path created: " + target_file)

	var dir_name, _ = filepath.Split(target_file)
	var to_path string = filepath.Join(dir_name, new_name)
	fmt.Println(GREEN + ">> " + RESET + "New path created: " + to_path)
	var _, err1 = os.Stat(to_path)
	if errors.Is(err1, fs.ErrNotExist) {
		var err error = os.Rename(target_file, to_path)

		if err != nil {
			fmt.Println(RED + ">> " + RESET + "A system error occured during file renaming.")
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		} else {
			fmt.Println(GREEN + ">> " + RESET + "File renaming was successful. (" + get_datentime() + ")")
			w.WriteHeader(http.StatusOK)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
	} else {
		fmt.Println(RED + ">> " + RESET + "A taken name was returned.")
		w.WriteHeader(http.StatusForbidden)
		fmt.Println(PINK + "-------------------------------------------------" + RESET)
		return
	}
}

func http_usesAuthQM(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	fmt.Println(CYAN + ">> " + RESET + "CLIENT: Do you have a password? (asked at: " + get_datentime() + ")")
	if get_settings().UseAuth {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusNotImplemented)
	}
	fmt.Println(PINK + "-------------------------------------------------" + RESET)
}
func http_yrAuthQM(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	var auth string = r.URL.Query().Get("auth")
	fmt.Println(CYAN + ">> " + RESET + "CLIENT: Is your password " + string([]rune(get_settings().Auth)[0]) + "***? (asked at: " + get_datentime() + ")")
	fmt.Println(PINK + "-------------------------------------------------" + RESET)

	if get_settings().UseAuth && auth == get_settings().Auth {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		w.WriteHeader(http.StatusGone)
		return
	}
}

func http_getSettings(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if get_settings().UseAuth {
		var auth string = r.URL.Query().Get("auth")
		if !(auth == get_settings().Auth) {
			fmt.Println(RED + ">> " + RESET + "Invalid authentication.")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
	}

	fmt.Println(CYAN + ">> " + RESET + "Attempt to get settings initiated. (" + get_datentime() + ")")

	json.NewEncoder(w).Encode(get_settings())
}

func http_setSettings(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if get_settings().UseAuth {
		var auth string = r.URL.Query().Get("auth")
		if !(auth == get_settings().Auth) {
			fmt.Println(RED + ">> " + RESET + "Invalid authentication.")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
	}
	var setting string = r.URL.Query().Get("setting")
	var arg string = r.URL.Query().Get("arg")

	switch setting {
	case "port":
		var config_json, err = os.Open(filepath.Join(filePath, "config.json"))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to open configuration file.")
			return
		}
		defer config_json.Close()
		var settings UserPreferences
		var err2 error = json.NewDecoder(config_json).Decode(&settings)
		if err2 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to read configuration file.")
			return
		}

		var result, err3 = strconv.Atoi(arg)
		if err3 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A server error occured while trying to convert a string to an integer. 🙄🙄")
			return
		}
		settings.Port = result

		var err4 error = json.NewEncoder(config_json).Encode(settings)
		if err4 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to write configuration file.")
			return
		}
	case "useAuth":
		var config_json, err = os.Open(filepath.Join(filePath, "config.json"))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to open configuration file.")
			return
		}
		defer config_json.Close()
		var settings UserPreferences
		var err2 error = json.NewDecoder(config_json).Decode(&settings)
		if err2 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to read configuration file.")
			return
		}

		if arg == "true" {
			settings.UseAuth = true
		} else {
			settings.UseAuth = false
		}

		var err3 error = json.NewEncoder(config_json).Encode(settings)
		if err3 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to write configuration file.")
			return
		}
	case "blacklist":
		var config_json, err = os.Open(filepath.Join(filePath, "config.json"))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to open configuration file.")
			return
		}
		defer config_json.Close()
		var settings UserPreferences
		var err2 error = json.NewDecoder(config_json).Decode(&settings)
		if err2 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to read configuration file.")
			return
		}

		settings.Blacklist = append(settings.Blacklist, arg)

		var err3 error = json.NewEncoder(config_json).Encode(settings)
		if err3 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to write configuration file.")
			return
		}
	case "authen":
		var config_json, err = os.Open(filepath.Join(filePath, "config.json"))
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to open configuration file.")
			return
		}
		defer config_json.Close()
		var settings UserPreferences
		var err2 error = json.NewDecoder(config_json).Decode(&settings)
		if err2 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to read configuration file.")
			return
		}

		settings.Auth = arg

		var err3 error = json.NewEncoder(config_json).Encode(settings)
		if err3 != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(RED + ">> " + RESET + "A system error occured while trying to write configuration file.")
			return
		}
	}
}

func http_uploadChunk(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if get_settings().UseAuth {
		var auth string = r.URL.Query().Get("auth")
		if !(auth == get_settings().Auth) {
			fmt.Println(RED + ">> " + RESET + "Invalid authentication.")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
	}
	var filename string = r.URL.Query().Get("filename")
	fmt.Println(CYAN + ">> " + RESET + "Attempt to upload files initiated. (" + get_datentime() + ")")

	r.Body = http.MaxBytesReader(w, r.Body, 10<<30)
	var target_path string = filepath.Join(filePath, "storage", filepath.Join(strings.Split(filename, "~/~")...))
	os.MkdirAll(filepath.Dir(target_path), 0755)

	var dest, err1 = os.OpenFile(target_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err1 != nil {
		fmt.Println(RED + ">> " + RESET + "A system error occured while trying to open the target file.")
		w.WriteHeader(http.StatusForbidden)
		return
	}
	defer dest.Close()
	var _, err2 = io.Copy(dest, r.Body)
	if err2 != nil {
		fmt.Println(RED + ">> " + RESET + "A system error occured while trying to write to the target file.")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	fmt.Println(GREEN + ">> " + RESET + "Successfully written all chunks! (" + get_datentime() + ")")
	fmt.Println(PINK + "-------------------------------------------------" + RESET)
}

func http_getContent(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if get_settings().UseAuth {
		var auth string = r.URL.Query().Get("auth")
		if !(auth == get_settings().Auth) {
			fmt.Println(RED + ">> " + RESET + "Invalid authentication.")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
	}
	var path string = r.URL.Query().Get("path")

	var parts []string = strings.Split(path, "~/~")
	path = filepath.Join(parts...)
	fmt.Println(CYAN + ">> " + RESET + "Working on: " + path)
	var fullpath string = filepath.Join(filePath, "storage", path)
	fmt.Println(GREEN + ">> " + RESET + "Target path created: " + fullpath)

	var file_content, err = os.ReadFile(fullpath)
	if err != nil {
		fmt.Println(RED + ">> " + RESET + "A system error occured while trying to read file.")
		w.WriteHeader(http.StatusForbidden)
		fmt.Println(PINK + "-------------------------------------------------" + RESET)
		return
	}
	fmt.Println(GREEN + ">> " + RESET + "Successfully read file.")
	var contentType string = http.DetectContentType(file_content[:512])
	fmt.Println(GREEN + ">> " + RESET + "File typed received: " + contentType)
	w.Header().Set("Content-Type", contentType)
	w.Write(file_content)
	fmt.Println(GREEN + ">> " + RESET + "File content returned! (" + get_datentime() + ")")
	fmt.Println(PINK + "-------------------------------------------------" + RESET)
}

func http_getFiles(w http.ResponseWriter, r *http.Request) {
	for _, ip := range get_settings().Blacklist {
		if host, _, _ := net.SplitHostPort(r.RemoteAddr); host == ip {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
	if get_settings().UseAuth {
		var auth string = r.URL.Query().Get("auth")
		if !(auth == get_settings().Auth) {
			fmt.Println(RED + ">> " + RESET + "Invalid authentication.")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
	}
	var search_dir string = r.URL.Query().Get("dir")
	fmt.Println(CYAN + ">> " + RESET + "Attempt to get file names initiated. (" + get_datentime() + ")")

	fmt.Println(CYAN + ">> " + RESET + "Directory returned: '" + search_dir + "'")
	var parts []string = strings.Split(search_dir, "~/~")
	search_dir = filepath.Clean(filepath.Join(parts...))
	fmt.Println(GREEN + ">> " + RESET + "Short path created: " + search_dir)

	var toReturn []DirEntry
	var files []os.DirEntry

	if search_dir == "" {
		var search_path string = filepath.Join(filePath, "storage")
		fmt.Println(GREEN + ">> " + RESET + "Full path created: " + search_path)
		var x, err = os.ReadDir(search_path)
		if err != nil {
			fmt.Println(RED + ">> " + RESET + "A system error occured during directory indexing.")
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
			return
		}
		files = x
	} else {
		var search_path string = filepath.Join(filePath, "storage", search_dir)
		fmt.Println(GREEN + ">> " + RESET + "Full path created: " + search_path)
		var x, err = os.ReadDir(search_path)
		if err != nil {
			fmt.Println(RED + ">> " + RESET + "A system error occured during directory indexing.")
			w.WriteHeader(http.StatusForbidden)
			fmt.Println(PINK + "-------------------------------------------------" + RESET)
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

	fmt.Println(GREEN + ">> " + RESET + "All files stored in a list.")

	json.NewEncoder(w).Encode(toReturn)
	fmt.Println(GREEN + ">> " + RESET + "Directory indexing was successful. (" + get_datentime() + ")")
	fmt.Println(PINK + "-------------------------------------------------" + RESET)
}

func main() {
	var afp, _ = filepath.Abs(filePath)
	var safp, _ = filepath.Split(afp)
	filePath = safp

	fmt.Println(BLUE + "Initializing " + PINK + "listener..." + RESET)
	var listener, err = net.Listen("tcp", ":"+strconv.Itoa(get_settings().Port))
	if err != nil {
		fmt.Println(RED+"ERROR while initializing listener:", err.Error())
		os.Exit(1)
	} else {
		fmt.Println(GREEN + "Listener created!" + RESET + " Listening at " + PINK + listener.Addr().String() + RESET)
	}

	fmt.Println(BLUE + "Starting " + PINK + "HTTP server..." + RESET)
	http.HandleFunc("/", http_main)
	http.HandleFunc("/explorer", http_explore)
	http.HandleFunc("/delete-file", http_deleteFile)
	http.HandleFunc("/get-files", http_getFiles)
	http.HandleFunc("/get-settings", http_getSettings)
	http.HandleFunc("/use-auth-qm", http_usesAuthQM)
	http.HandleFunc("/yr-auth-qm", http_yrAuthQM)
	http.HandleFunc("/set-settings", http_setSettings)
	http.HandleFunc("/get-content", http_getContent)
	http.HandleFunc("/rename-file", http_renameFile)
	http.HandleFunc("/upload-chunk", http_uploadChunk)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	fmt.Print("\033c")
	fmt.Println(GREEN + "Server initiated!" + RESET + " Serving at " + PINK + listener.Addr().String() + RESET)
	fmt.Println("Press " + PINK + "<Ctrl-C>" + RESET + " to stop the server.")
	fmt.Println(PINK + "-------------------------------------------------" + RESET)

	http.Serve(listener, nil)
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type DirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var _, filePath, _, _ = runtime.Caller(0);
var (
	RED string = "\033[31m"
	BLUE string = "\033[34m"
	CYAN string = "\033[36m"
	PINK string = "\033[35m"
	GREEN string = "\033[32m"
	RESET string = "\033[0m"
)

// settings
var (
	user_auth string = "likoliko12"
)

func http_main(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir(filepath.Join(filePath, "web"))).ServeHTTP(w, r);
}
func http_explore(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/explorer", http.FileServer(http.Dir(filepath.Join(filePath, "web", "explorer")))).ServeHTTP(w, r);
}
func http_verify(w http.ResponseWriter, r *http.Request) {
	var auth string = r.URL.Query().Get("auth");
	fmt.Println(CYAN+">> "+RESET+"Log-in attempt initiated.");
	if (auth == user_auth) {
		fmt.Println(GREEN+">> "+RESET+"Log-in attempt successful.");
		w.WriteHeader(http.StatusOK);
	} else {
		fmt.Println(RED+">> "+RESET+"Log-in attempt failed.");
		w.WriteHeader(http.StatusForbidden);
	}

	fmt.Println(PINK+"-------------------------------------------------"+RESET);
}
func http_deleteFile(w http.ResponseWriter, r *http.Request) {
	var auth string = r.URL.Query().Get("auth")
	var file_td string = r.URL.Query().Get("name");
	fmt.Println(CYAN+">> "+RESET+"Attempt to delete a file initiated.");

	if auth == "" || !(auth == user_auth) {
		fmt.Println(RED+">> "+RESET+"Invalid authentication.");
		w.WriteHeader(http.StatusBadRequest);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	}
	
	if file_td == "" {
		fmt.Println(RED+">> "+RESET+"No target returned.")
		w.WriteHeader(http.StatusNotFound);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	}
	var parts []string = strings.Split(file_td, "~/~");
	file_td = filepath.Clean(filepath.Join(parts...));
	var ts_filePath string = filepath.Join(filePath, "storage", file_td);
	fmt.Println(GREEN+">> "+RESET+"Path created: "+ts_filePath);

	var _, err = os.Stat(ts_filePath);
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println(RED+">> "+RESET+"Nonexistent file returned.")
		w.WriteHeader(http.StatusNotFound);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	}

	if os.RemoveAll(ts_filePath) != nil {
		fmt.Println(RED+">> "+RESET+"A system error occured during file removal.")
		w.WriteHeader(http.StatusForbidden);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	} else {
		fmt.Println(GREEN+">> "+RESET+"File removal successful.")
		w.WriteHeader(http.StatusFound);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	}
}

func http_renameFile(w http.ResponseWriter, r *http.Request) {
	var auth string = r.URL.Query().Get("auth");
	var path string = r.URL.Query().Get("path");
	var new_name string = r.URL.Query().Get("newname");
	fmt.Println(CYAN+">> "+RESET+"Attempt to rename a file initiated.");
	
	if auth == "" || !(auth == user_auth) {
		fmt.Println(RED+">> "+RESET+"Invalid authentication.");
		w.WriteHeader(http.StatusBadRequest);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	}

	if path == "" || new_name == "" || (path == "" && new_name == "") {
		fmt.Println(RED+">> "+RESET+"No target returned.")
		w.WriteHeader(http.StatusBadRequest);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	}

	var parts []string = strings.Split(path, "~/~");
	path = filepath.Join(parts...);
	var target_file string = filepath.Join(filePath, "storage", path);
	fmt.Println(GREEN+">> "+RESET+"Path created: "+target_file);

	var dir_name, _ = filepath.Split(target_file);
	var to_path string = filepath.Join(dir_name, new_name);
	fmt.Println(GREEN+">> "+RESET+"New path created: "+to_path);
	var _, err1 = os.Stat(to_path);
	if errors.Is(err1, fs.ErrNotExist) {
		var err error = os.Rename(target_file, to_path);

		if err != nil {
			fmt.Println(RED+">> "+RESET+"A system error occured during file renaming.")
			w.WriteHeader(http.StatusForbidden);
			fmt.Println(PINK+"-------------------------------------------------"+RESET);
			return;
		} else {
			fmt.Println(GREEN+">> "+RESET+"File renaming was successful.")
			w.WriteHeader(http.StatusOK);
			fmt.Println(PINK+"-------------------------------------------------"+RESET);
			return;
		}
	} else {
		fmt.Println(RED+">> "+RESET+"A taken name was returned.");
		w.WriteHeader(http.StatusForbidden);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	}
}

func http_getFiles(w http.ResponseWriter, r *http.Request) {
	var auth string = r.URL.Query().Get("auth");
	var search_dir string = r.URL.Query().Get("dir");
	fmt.Println(CYAN+">> "+RESET+"Attempt to get file names initiated.");

	if auth == "" || !(auth == user_auth) {
		fmt.Println(RED+">> "+RESET+"Invalid authentication.");
		w.WriteHeader(http.StatusBadRequest);
		fmt.Println(PINK+"-------------------------------------------------"+RESET);
		return;
	}

	fmt.Println(CYAN+">> "+RESET+"Directory returned: "+search_dir);
	var parts []string = strings.Split(search_dir, "~/~");
	search_dir = filepath.Clean(filepath.Join(parts...));
	fmt.Println(GREEN+">> "+RESET+"Short path created: "+search_dir);

	var toReturn []DirEntry;
	var files []os.DirEntry;

	if search_dir == "" {
		var search_path string = filepath.Join(filePath, "storage");
		fmt.Println(GREEN+">> "+RESET+"Full path created: "+search_path);
		var x, err = os.ReadDir(search_path);
		if err != nil {
			fmt.Println(RED+">> "+RESET+"A system error occured during directory indexing.");
			w.WriteHeader(http.StatusForbidden);
			fmt.Println(PINK+"-------------------------------------------------"+RESET);
			return;
		}
		files = x;
	} else {
		var search_path string = filepath.Join(filePath, "storage", search_dir);
		fmt.Println(GREEN+">> "+RESET+"Full path created: "+search_path);
		var x, err = os.ReadDir(search_path);
		if err != nil {
			fmt.Println(RED+">> "+RESET+"A system error occured during directory indexing.");
			w.WriteHeader(http.StatusForbidden);
			fmt.Println(PINK+"-------------------------------------------------"+RESET);
			return;
		}
		files = x;
	}

	for _, v := range files {
		var file DirEntry;
		file.Name = v.Name();
		if v.Type().IsDir() {
			file.Type = "d";
		} else { file.Type = "f"; }

		toReturn = append(toReturn, file);
	}
	fmt.Println(GREEN+">> "+RESET+"All files stored in a list.");

	json.NewEncoder(w).Encode(toReturn);
	fmt.Println(GREEN+">> "+RESET+"Directory indexing was successful.");
	fmt.Println(PINK+"-------------------------------------------------"+RESET);
}

func main() {
	var afp, _ = filepath.Abs(filePath);
	var safp, _ = filepath.Split(afp);
	filePath = safp;

	fmt.Println(BLUE+"Initializing "+PINK+"listener..."+RESET);
	var listener, err = net.Listen("tcp", ":0");
	if err != nil {
		fmt.Println(RED+"ERROR while initializing listener:", err.Error());
		os.Exit(1);
	} else {
		fmt.Println(GREEN+"Listener created!"+RESET+" Listening at "+PINK+listener.Addr().String()+RESET);
	}

	fmt.Println(BLUE+"Starting "+PINK+"HTTP server..."+RESET);
	http.HandleFunc("/", http_main);
	http.HandleFunc("/explorer", http_explore);
	http.HandleFunc("/delete-file", http_deleteFile);
	http.HandleFunc("/get-files", http_getFiles);
	http.HandleFunc("/rename-file", http_renameFile);
	http.HandleFunc("/verify", http_verify);
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent);
	});
	fmt.Print("\033c");
	fmt.Println(GREEN+"Server initiated!"+RESET+" Serving at "+PINK+listener.Addr().String()+RESET);
	fmt.Println("Press "+PINK+"<Ctrl-C>"+RESET+" to stop the server.");
	fmt.Println(PINK+"-------------------------------------------------"+RESET);

	http.Serve(listener, nil);
}


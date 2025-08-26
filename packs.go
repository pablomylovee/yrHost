package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/facette/natsort"
	"github.com/gofiber/fiber/v2"
)

func http_createPack(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden);
	}
	log(ATTEMPT, "Attempt to create a pack initiated", false);

	var serviceParent string;
	switch c.Query("service") {
	case "files": serviceParent = filepath.Join(yf_savePath, c.Query("username"));
	case "sound": serviceParent = filepath.Join(ys_savePath, c.Query("username"));
	default: return c.SendStatus(fiber.StatusBadRequest);
	}
	log(COMPLETE, "Received service!", false);

	log(ATTEMPT, "Creating directory for pack", false);
	os.Mkdir(filepath.Join(serviceParent, "packs", c.Query("name")), 0755);
	log(COMPLETE, "Successfully created pack!", true);
	return c.SendStatus(fiber.StatusOK);
}

func http_createEntry(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden);
	}
	log(ATTEMPT, "Attempt to create an entry in a pack initiated.", false);

	var serviceParent string;
	switch c.Query("service") {
	case "files": serviceParent = filepath.Join(yf_savePath, c.Query("username"));
	case "sound": serviceParent = filepath.Join(ys_savePath, c.Query("username"));
	default: return c.SendStatus(fiber.StatusBadRequest);
	}
	log(COMPLETE, "Received service!", false);

	log(ATTEMPT, "Creating entry", false);
	os.Mkdir(filepath.Join(serviceParent, "packs", c.Query("name"), c.Query("id")), 0755);
	var file, err = os.Create(filepath.Join(serviceParent, "packs", c.Query("name"), c.Query("id"), "config.json"));
	if err != nil {
		log(ERROR, "A system error occured while trying to create entry.", true);
		return c.SendStatus(fiber.StatusInternalServerError);
	}
	file.Write(c.Body());
	log(COMPLETE, "Successfully created entry!", true);
	return c.SendStatus(fiber.StatusOK);
}

func http_appendChunk(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden);
	}
	log(ATTEMPT, "Attempt to append chunk to a pack entry initiated.", false);


	var serviceParent string;
	switch c.Query("service") {
	case "files": serviceParent = filepath.Join(yf_savePath, c.Query("username"));
	case "sound": serviceParent = filepath.Join(ys_savePath, c.Query("username"));
	default: return c.SendStatus(fiber.StatusBadRequest);
	}
	log(COMPLETE, "Received service!", false);

	log(ATTEMPT, "Creating space for chunk", false);
	var chunk, err = os.Create(filepath.Join(serviceParent, "packs", c.Query("name"), c.Query("id"), fmt.Sprintf("%d.part", c.QueryInt("part"))));
	if err != nil {
		log(ERROR, "A system error occured while trying to create space.", true);
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	log(COMPLETE, "Created!", false);
	log(ATTEMPT, "Writing to chunk", false);
	var _, err1 = chunk.Write(c.Body());
	if err1 != nil {
		log(ERROR, "A system error occured while trying to write to chunk.", true);
		return c.SendStatus(fiber.StatusInternalServerError);
	}

	log(COMPLETE, "Successfully appended chunk to pack entry!", true);
	return c.SendStatus(fiber.StatusOK);
}

func http_assemblePack(c *fiber.Ctx) error {
	if !check_allowed(c) || !check_auth(c) {
		return c.SendStatus(fiber.StatusForbidden);
	}
	log(ATTEMPT, "Attempt to assemble pack initiated.", false);


	var serviceParent string;
	switch c.Query("service") {
	case "files": serviceParent = filepath.Join(yf_savePath, c.Query("username"));
	case "sound": serviceParent = filepath.Join(ys_savePath, c.Query("username"));
	default: return c.SendStatus(fiber.StatusBadRequest);
	}
	log(COMPLETE, "Received service!", false);

	var packPath string = filepath.Join(serviceParent, "packs", c.Query("name"));
	var entries, err = os.ReadDir(packPath);
	if err != nil {
		log(ERROR, "A system error occured while trying to get entries for pack.", true);
		return c.SendStatus(fiber.StatusInternalServerError);
	}

	for _, entry := range entries {
		if !entry.Type().IsDir() {
			continue
		}
		var entryPath string = filepath.Join(packPath, entry.Name());
		var config, err1 = os.ReadFile(filepath.Join(entryPath, "config.json"));
		var chunks, err5 = os.ReadDir(entryPath);
		if err5 != nil {
			log(ERROR, "A system error occured while trying to get all chunks from entry.", true);
			return c.SendStatus(fiber.StatusInternalServerError);
		}
		if err1 != nil {
			log(ERROR, "A system error occured while trying to read configuration.", true);
			return c.SendStatus(fiber.StatusInternalServerError);
		}
		var options struct{
			RelativePath string `json:"relative-path"`
		}
		var err2 error = json.Unmarshal(config, &options);
		if err2 != nil {
			log(ERROR, "An error occured while trying to read configuration", true);
			return c.SendStatus(fiber.StatusInternalServerError);
		}

		var chunkNames []string;
		for _, chunk := range chunks {
			if chunk.Name() == "config.json" {
				continue
			}
			chunkNames = append(chunkNames, chunk.Name());
		}
		natsort.Sort(chunkNames);

		var parts []string = strings.Split(options.RelativePath, string(os.PathSeparator));
		os.MkdirAll(filepath.Dir(filepath.Join(serviceParent, "files", filepath.Join(parts...))), 0755);
		var file, err3 = os.OpenFile(filepath.Join(serviceParent, "files", filepath.Join(parts...)), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644);
		if err3 != nil {
			log(ERROR, "A system error occured while trying to open target entry file.", true);
			return c.SendStatus(fiber.StatusInternalServerError);
		}
		for _, chunkName := range chunkNames {
			var chunkContent, err4 = os.ReadFile(filepath.Join(entryPath, chunkName));
			if err4 != nil {
				log(ERROR, "A system error occured while trying to read chunk.", true);
				return c.SendStatus(fiber.StatusInternalServerError);
			}
			file.Write(chunkContent);
		}
		file.Close();
	}
	os.RemoveAll(packPath);
	return c.SendStatus(fiber.StatusOK);
}


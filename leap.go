package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	bulletinURL     = "https://datacenter.iers.org/data/latestVersion/bulletinC.txt"
	apiURL          = "https://leap.webclock.io/leap.json"
	gitHubSourceURL = "https://github.com/nixigaj/leap/blob/pages/leap.json"
)

type config struct {
	GotifyURL   string `json:"gotify_url"`
	GotifyToken string `json:"gotify_token"`
	FilePath    string `json:"file_path"`
}

func main() {
	cfg := getConfig()
	if err := initialCheck(cfg); err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Printf("Starting monitoring of %s", bulletinURL)

	for {
		select {
		case <-ticker.C:
			if err := checkForUpdates(cfg); err != nil {
				log.Printf("Error checking for updates: %v", err)
			}
		}
	}
}

func getConfig() config {
	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
	flag.Parse()

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg config
	if err := json.Unmarshal(configFile, &cfg); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	if cfg.GotifyURL == "" || cfg.GotifyToken == "" {
		log.Fatal("Gotify URL and token are required in config file")
	}

	if cfg.FilePath == "" {
		cfg.FilePath = "bulletinC.txt"
	}

	return cfg
}

func initialCheck(cfg config) error {
	_, err := os.Stat(cfg.FilePath)
	if os.IsNotExist(err) {
		return fetchAndSave(cfg.FilePath)
	} else if err != nil {
		return fmt.Errorf("error checking file: %v", err)
	}
	return checkForUpdates(cfg)
}

func fetchAndSave(filePath string) error {
	content, err := fetchURL(bulletinURL)
	if err != nil {
		return fmt.Errorf("error fetching bulletin: %v", err)
	}

	return os.WriteFile(filePath, content, 0644)
}

func fetchURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %v", err)
	}

	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("failed to close response body: %v", err)
	}

	return data, nil
}

func checkForUpdates(cfg config) error {
	newContent, err := fetchURL(bulletinURL)
	if err != nil {
		return fmt.Errorf("error fetching bulletin: %v", err)
	}

	existingContent, err := os.ReadFile(cfg.FilePath)
	if err != nil {
		return fmt.Errorf("error reading existing file: %v", err)
	}

	if !bytes.Equal(existingContent, newContent) {
		message, err := generateNotificationMessage(newContent)
		if err != nil {
			return fmt.Errorf("error generating notification message: %v", err)
		}

		if err := sendNotification(cfg, message); err != nil {
			return fmt.Errorf("error sending notification: %v", err)
		}

		if err := os.WriteFile(cfg.FilePath, newContent, 0644); err != nil {
			return fmt.Errorf("error updating file: %v", err)
		}

		log.Printf("Bulletin updated and notification sent")
	}

	return nil
}

func generateNotificationMessage(newContent []byte) (string, error) {
	apiJSON, err := fetchURL(apiURL)
	if err != nil {
		return "", fmt.Errorf("error fetching API: %v", err)
	}

	message := "============\nNew content:\n============\n"
	message += strings.TrimSpace(string(newContent)) + "\n"
	message += "\n============\nAPI content:\n============\n"
	message += string(apiJSON) + "\n"
	message += "\nGitHub source URL: " + gitHubSourceURL

	return message, nil
}

func sendNotification(cfg config, message string) error {
	url := fmt.Sprintf("%s/message?token=%s", cfg.GotifyURL, cfg.GotifyToken)

	resp, err := http.PostForm(url, map[string][]string{
		"title":    {"IERS Bulletin C Update"},
		"message":  {message},
		"priority": {"5"},
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from Gotify: %d", resp.StatusCode)
	}

	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("failed to close response body: %v", err)
	}

	return nil
}

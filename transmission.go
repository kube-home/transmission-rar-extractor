package main

import (
	"fmt"
	"github.com/tubbebubbe/transmission"
	"os"
)

// connection Creates a RPC connection towards Transmission
func connection() transmission.TransmissionClient {
	serverUrl := os.Getenv("TRANSMISSION_HOST")
	if serverUrl == "" {
		serverUrl = "http://127.0.0.1:9091"
	}
	username := os.Getenv("TRANSMISSION_USER")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("TRANSMISSION_PASS")

	transmissionbt := transmission.New(serverUrl, username, password)

	return transmissionbt
}

// Scan connects to a transmission RPC and gets all torrents and their dir path
// Then sends them through the channel
func Scan(jobs chan<- string) {
	transmissionbt := connection()
	torrents, err := transmissionbt.GetTorrents()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, torrent := range torrents {
		fmt.Println("Torrent status:", torrent.Status)
		if torrent.Status == 6 {
			entityPath := fmt.Sprintf("%s/%s", torrent.DownloadDir, torrent.Name)
			if _, err := os.Stat(fmt.Sprintf("%s/norar", entityPath)); err == nil {
				fmt.Println("Skipping...")
				continue
			} else if os.IsNotExist(err) {
				jobs <- entityPath
			}
		}
	}
}

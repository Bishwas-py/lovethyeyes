package main

import (
	"log"
	"os/exec"
)

func notification(appName string, title string, text string, iconPath string) {
	cmd := exec.Command("notify-send", "-i", iconPath, title, text, "-a", appName)
	err := cmd.Run()
	if err != nil {
		return
	}
	log.Println("Error: ", err)
}

package main

import (
	"log"
	"os/exec"
)

func SnoreToast(title string, message string, imagePath string) error {
	log.Println(title, ":", message)
	var err error
	if len(imagePath) <= 0 {
		err = exec.Command("SnoreToast.exe", "-t", title, "-m", message, "-s", "Notification.IM").Run()
	} else {
		err = exec.Command("SnoreToast.exe", "-t", title, "-m", message, "-p", imagePath, "-s", "Notification.IM").Run()
	}
	if err != nil {
		return err
	}
	return nil
}

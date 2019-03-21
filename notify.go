package main

import (
	"log"
	"os/exec"
)

func SnoreToast(title string, message string) error {
	log.Println(title, ":", message)
	err := exec.Command("SnoreToast.exe", "-t", title, "-m", message).Run()
	if err != nil {
		return err
	}
	return nil
}

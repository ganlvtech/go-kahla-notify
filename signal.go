package main

import (
	"os"
	"os/signal"
)

func WaitForCtrlC() {
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
}

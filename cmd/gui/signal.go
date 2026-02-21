package main

import (
	"os"
	"os/signal"
	"syscall"
)

func signalNotifyImpl(ch chan os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}

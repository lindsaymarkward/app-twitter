package main

// app-twitter Ninja Sphere app for Twitter (notifications)
// Lindsay Ward, July 2015 - https://github.com/lindsaymarkward/app-twitter

import (
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/support"
)

var log = logger.GetLogger(info.Name)

func main() {
	app := &TwitterApp{}

	err := app.Init(info)
	if err != nil {
		app.Log.Fatalf("failed to initialize app: %v", err)
	}

	err = app.Export(app)
	if err != nil {
		app.Log.Fatalf("failed to export app: %v", err)
	}

	support.WaitUntilSignal()
}

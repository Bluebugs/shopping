package main

import (
	"crypto/ed25519"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"github.com/fynelabs/fyneselfupdate"
	"github.com/fynelabs/selfupdate"
)

// selfManage turns on automatic updates
func selfManage(a fyne.App, w fyne.Window) {
	publicKey := ed25519.PublicKey{107, 151, 11, 16, 121, 196, 121, 53, 75, 255, 139, 217, 169, 149, 103, 159, 194, 69, 216, 72, 234, 214, 205, 26, 152, 37, 67, 70, 239, 131, 54, 106}

	// The public key above matches the signature of the below file served by our CDN
	httpSource := selfupdate.NewHTTPSource(nil, "https://geoffrey-artefacts.fynelabs.com/self-update/7b/7b003726-14e0-4f1a-82a5-900217d75528/{{.OS}}-{{.Arch}}/{{.Executable}}{{.Ext}}")

	config := fyneselfupdate.NewConfigWithTimeout(a, w, time.Minute, httpSource, selfupdate.Schedule{FetchOnStart: true, Interval: time.Hour * 12}, publicKey)

	_, err := selfupdate.Manage(config)
	if err != nil {
		log.Println("Error while setting up update manager: ", err)
		return
	}
}

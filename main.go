package main

import (
	"fmt"
	"github.com/Bishwas-py/notify"
	"github.com/godbus/dbus/v5"
	"log"
	"path"
	"runtime"
	"time"
)

const AppID = "Love Thy Eyes"

var AppIcon string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	AppIcon = path.Join(path.Dir(filename), "logo.png")
}

func main() {
	hints := map[string]dbus.Variant{
		"body-markup":     dbus.MakeVariant(true), // Enable markup in body
		"body-hyperlinks": dbus.MakeVariant(true), // Enable hyperlinks specifically
	}

	notification := notify.Notification{
		Title:   "Now, take a break!",
		AppIcon: AppIcon,
		AppID:   AppID,
		Body:    "Go out, see something beautiful, something far away, a mountain, a river, a forest or a sea. It will help you to relax your eyes.",
		Actions: notify.Actions{
			{
				Title: "Hello",
				Trigger: func() {
					log.Println("Hello")
				},
			},
		},
		Hints:   hints,
		Timeout: int(10 * time.Second),
	}

	_, s := notification.Actions.Results()
	fmt.Printf("%v", s)

	_, _ = notification.Trigger()
}

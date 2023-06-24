package main

import (
	"embed"
	"fmt"

	"github.com/kaptinlin/go-i18n"
	"github.com/pelletier/go-toml/v2"
)

//go:embed locales/*
var localesFs embed.FS

func main() {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
		i18n.WithUnmarshaler(toml.Unmarshal),
	)

	err := bundle.LoadFS(localesFs, "locales/*.toml")
	if err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")

	// Output: Hello, world
	fmt.Println(localizer.Get("hello_world"))

	// Output: Hello, John
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))

	// Output: Message
	fmt.Println(localizer.Get("message", i18n.Vars{
		"count": 1,
	}))

	// Output: Messages
	fmt.Println(localizer.Get("message", i18n.Vars{
		"count": 2,
	}))

	// Output: No messages
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 0,
	}))

	// Output: 1 message
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 1,
	}))

	// Output: 2 messages
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 2,
	}))
}

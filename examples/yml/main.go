package main

import (
	"embed"
	"fmt"

	"github.com/goccy/go-yaml"

	"github.com/kaptinlin/go-i18n"
)

//go:embed locales/*
var localesFs embed.FS

func main() {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
		i18n.WithUnmarshaler(yaml.Unmarshal),
	)

	err := bundle.LoadFS(localesFs, "locales/*.yml")
	if err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")

	fmt.Println(localizer.Get("hello_world"))
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))
	fmt.Println(localizer.Get("message", i18n.Vars{
		"count": 1,
	}))
	fmt.Println(localizer.Get("message", i18n.Vars{
		"count": 2,
	}))
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 0,
	}))
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 1,
	}))
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 2,
	}))
}

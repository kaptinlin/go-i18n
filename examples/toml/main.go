package main

import (
	"embed"
	"fmt"

	"github.com/pelletier/go-toml/v2"

	"github.com/kaptinlin/go-i18n"
)

//go:embed locales/*
var localesFs embed.FS

func main() {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
		i18n.WithUnmarshaler(toml.Unmarshal),
	)

	if err := bundle.LoadFS(localesFs, "locales/*.toml"); err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")
	print := func(key string, vars i18n.Vars) {
		fmt.Println(localizer.Get(key, vars))
	}

	fmt.Println(localizer.Get("hello_world"))
	print("hello_name", i18n.Vars{"name": "John"})
	print("message", i18n.Vars{"count": 1})
	print("message", i18n.Vars{"count": 2})
	print("message_with_number", i18n.Vars{"count": 0})
	print("message_with_number", i18n.Vars{"count": 1})
	print("message_with_number", i18n.Vars{"count": 2})
}

package main

import (
	"embed"
	"fmt"

	"github.com/kaptinlin/go-i18n"
)

//go:embed locales/*
var localesFs embed.FS

func main() {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
	)

	if err := bundle.LoadFS(localesFs, "locales/*.json"); err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")

	fmt.Println(localizer.Get("hello_world"))
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))
}

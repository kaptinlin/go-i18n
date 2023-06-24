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

	err := bundle.LoadFS(localesFs, "locales/*.json")
	if err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")

	// Output: 你好，世界
	fmt.Println(localizer.Get("hello_world"))

	// Output: 你好, John
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))
}

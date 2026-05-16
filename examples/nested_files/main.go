package main

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/kaptinlin/go-i18n"
)

//go:embed locales/*
var localesFs embed.FS

func main() {
	run(localesFs, "*/*.json", "*/*/*.json")
}

func run(fsys fs.FS, patterns ...string) {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
	)

	if err := bundle.LoadFS(fsys, patterns...); err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("zh-Hans")

	fmt.Println(localizer.Get("hello_world"))
	fmt.Println(localizer.Get("message.hi", i18n.Vars{
		"name": "John",
	}))
}

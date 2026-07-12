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
	run(localesFs, "locales/*.json")
}

func run(fsys fs.FS, patterns ...string) {
	bundle, err := i18n.NewBundle("en",
		i18n.WithLocales("zh-Hans"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := bundle.LoadFS(fsys, patterns...); err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")

	fmt.Println(localizer.Get("hello_world"))
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))
}

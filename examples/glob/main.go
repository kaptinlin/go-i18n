package main

import (
	"fmt"

	"github.com/kaptinlin/go-i18n"
)

func main() {
	run("./locales/*.json")
}

func run(patterns ...string) {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
	)

	if err := bundle.LoadGlob(patterns...); err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("zh-Hans")

	fmt.Println(localizer.Get("hello_world"))
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))
}

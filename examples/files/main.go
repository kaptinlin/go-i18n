package main

import (
	"fmt"

	"github.com/kaptinlin/go-i18n"
)

func main() {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
	)

	err := bundle.LoadFiles("./locales/zh-Hans.json", "./locales/en.json")
	if err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("zh-Hans")

	// Output: 你好，世界
	fmt.Println(localizer.Get("hello_world"))

	// Output: 你好, John
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))
}

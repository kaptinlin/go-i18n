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

	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {
			"Hello world":   "Hello world",
			"Hello, {name}": "Hello, {name}",
			"Hello, %s":     "Hello, %s",
		},
		"zh-Hans": {
			"Hello world":   "你好, 世界",
			"Hello, {name}": "你好, {name}",
			"Hello, %s":     "你好, %s",
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("zh-Hans")

	fmt.Println(localizer.Get("Hello world"))
	fmt.Println(localizer.Get("Hello, {name}", i18n.Vars{
		"name": "John",
	}))
	fmt.Println(fmt.Sprintf(localizer.Get("Hello, %s"), "Alice"))
}

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
			"hello_world":  "Hello, world",
			"hello_name":   "Hello, {name}",
			"hello_string": "Hello, %s",
		},
		"zh-Hans": {
			"hello_world":  "你好, 世界",
			"hello_name":   "你好, {name}",
			"hello_string": "你好, %s",
		},
	})
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

	// Output: 你好，Alice
	fmt.Println(localizer.Getf("hello_string", "Alice"))
}

package main

import (
	"fmt"

	"github.com/kaptinlin/go-i18n"
)

func main() {
	run("./locales/*.json")
}

func run(patterns ...string) {
	bundle, err := i18n.NewBundle("en",
		i18n.WithLocales("zh-Hans"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := bundle.LoadGlob(patterns...); err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("zh-Hans")

	fmt.Println(localizer.Get("hello_world"))
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))
}

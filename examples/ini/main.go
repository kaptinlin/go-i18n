package main

import (
	"embed"
	"fmt"

	"github.com/kaptinlin/go-i18n"
	"gopkg.in/ini.v1"
)

//go:embed locales/*
var localesFs embed.FS

func unmarshalINI(data []byte, v any) error {
	f, err := ini.LoadSources(ini.LoadOptions{
		SpaceBeforeInlineComment: true,
		IgnoreInlineComment:      true,
	}, data)
	if err != nil {
		return err
	}

	m := *v.(*map[string]string)

	// Includes the ini.DefaultSection which has the root keys too.
	// We don't have to iterate to each section to find the subsection,
	// the Sections() returns all sections, sub-sections are separated by dot '.'
	// and we match the dot with a section on the translate function, so we just save the values as they are,
	// so we don't have to do section lookup on every translate call.
	for _, section := range f.Sections() {
		keyPrefix := ""
		if name := section.Name(); name != ini.DefaultSection {
			keyPrefix = name + "."
		}

		for _, key := range section.Keys() {
			m[keyPrefix+key.Name()] = key.Value()
		}
	}

	return nil
}

func main() {
	bundle := i18n.NewBundle(
		i18n.WithDefaultLocale("en"),
		i18n.WithLocales("en", "zh-Hans"),
		i18n.WithUnmarshaler(unmarshalINI),
	)

	err := bundle.LoadFS(localesFs, "locales/*.ini")
	if err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")

	// Output: Hello, world
	fmt.Println(localizer.Get("hello_world"))

	// Output: Hello, John
	fmt.Println(localizer.Get("hello_name", i18n.Vars{
		"name": "John",
	}))

	// Output: Message
	fmt.Println(localizer.Get("message", i18n.Vars{
		"count": 1,
	}))

	// Output: Messages
	fmt.Println(localizer.Get("message", i18n.Vars{
		"count": 2,
	}))

	// Output: No messages
	fmt.Println(localizer.Get("message.with_number", i18n.Vars{
		"count": 0,
	}))

	// Output: 1 message
	fmt.Println(localizer.Get("message.with_number", i18n.Vars{
		"count": 1,
	}))

	// Output: 2 messages
	fmt.Println(localizer.Get("message.with_number", i18n.Vars{
		"count": 2,
	}))
}

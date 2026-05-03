package main

import (
	"embed"
	"fmt"

	"gopkg.in/ini.v1"

	"github.com/kaptinlin/go-i18n"
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

	// Flatten section prefixes once so lookups can use the final dotted keys directly.
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

	if err := bundle.LoadFS(localesFs, "locales/*.ini"); err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")
	print := func(key string, vars i18n.Vars) {
		fmt.Println(localizer.Get(key, vars))
	}

	fmt.Println(localizer.Get("hello_world"))
	print("hello_name", i18n.Vars{"name": "John"})
	print("message", i18n.Vars{"count": 1})
	print("message", i18n.Vars{"count": 2})
	print("message.with_number", i18n.Vars{"count": 0})
	print("message.with_number", i18n.Vars{"count": 1})
	print("message.with_number", i18n.Vars{"count": 2})
}

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

	// Load translations
	err := bundle.LoadMessages(map[string]map[string]string{
		"en": {
			"hello":   "Hello, {name}!",
			"goodbye": "Goodbye, {name}!",
		},
		"zh-Hans": {
			"hello": "你好，{name}！",
			// Note: "goodbye" is not translated in zh-Hans
		},
	})
	if err != nil {
		panic(err)
	}

	localizer := bundle.NewLocalizer("zh-Hans")

	fmt.Println("=== Lookup Examples ===")

	// Example 1: Direct hit in requested locale
	r := localizer.Lookup("hello", i18n.Vars{"name": "World"})
	printResult("hello", r)

	// Example 2: Fallback to default locale
	r = localizer.Lookup("goodbye", i18n.Vars{"name": "World"})
	printResult("goodbye", r)

	// Example 3: Translation not found anywhere
	r = localizer.Lookup("unknown_key")
	printResult("unknown_key", r)

	fmt.Println("\n=== Detecting Fallback vs Direct Hit ===")

	r = localizer.Lookup("hello", i18n.Vars{"name": "World"})
	switch {
	case !r.Found:
		fmt.Printf("  %q: NOT FOUND\n", "hello")
	case r.Locale != localizer.Locale():
		fmt.Printf("  %q: fallback from %s\n", "hello", r.Locale)
	default:
		fmt.Printf("  %q: direct hit in %s\n", "hello", r.Locale)
	}

	r = localizer.Lookup("goodbye", i18n.Vars{"name": "World"})
	switch {
	case !r.Found:
		fmt.Printf("  %q: NOT FOUND\n", "goodbye")
	case r.Locale != localizer.Locale():
		fmt.Printf("  %q: fallback from %s\n", "goodbye", r.Locale)
	default:
		fmt.Printf("  %q: direct hit in %s\n", "goodbye", r.Locale)
	}

	fmt.Println("\n=== Context Disambiguation ===")

	err = bundle.LoadMessages(map[string]map[string]string{
		"zh-Hans": {
			"Post <verb>": "发布文章",
			"Post <noun>": "文章",
		},
	})
	if err != nil {
		panic(err)
	}

	// GetX for daily use
	fmt.Printf("  GetX verb: %s\n", localizer.GetX("Post", "verb"))
	fmt.Printf("  GetX noun: %s\n", localizer.GetX("Post", "noun"))

	// Lookup for details
	r = localizer.Lookup("Post <verb>")
	printResult("Post <verb>", r)
}

func printResult(key string, r i18n.TranslationResult) {
	fmt.Printf("Key: %q\n", key)
	fmt.Printf("  Text:   %q\n", r.Text)
	fmt.Printf("  Locale: %q\n", r.Locale)
	fmt.Printf("  Found:  %v\n", r.Found)
	fmt.Println()
}

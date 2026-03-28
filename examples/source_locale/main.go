package main

import (
	"errors"
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

	fmt.Println("=== GetWithLocale Examples ===")

	// Example 1: Successful translation lookup
	text, locale, err := localizer.GetWithLocale("hello", i18n.Vars{"name": "World"})
	printResult("hello", text, locale, err)

	// Example 2: Missing translation (falls back to default locale)
	text, locale, err = localizer.GetWithLocale("goodbye", i18n.Vars{"name": "World"})
	printResult("goodbye", text, locale, err)

	// Example 3: Translation not found anywhere
	text, locale, err = localizer.GetWithLocale("unknown_key")
	printResult("unknown_key", text, locale, err)

	fmt.Println("\n=== GetXWithLocale Examples ===")

	// Add context-based translations
	err = bundle.LoadMessages(map[string]map[string]string{
		"zh-Hans": {
			"Post <verb>": "发布文章",
			"Post <noun>": "文章",
		},
	})
	if err != nil {
		panic(err)
	}

	text, locale, err = localizer.GetXWithLocale("Post", "verb")
	printResult("Post (verb)", text, locale, err)

	text, locale, err = localizer.GetXWithLocale("Post", "noun")
	printResult("Post (noun)", text, locale, err)

	fmt.Println("\n=== GetfWithLocale Examples ===")

	// Note: GetfWithLocale uses fmt.Sprintf-style formatting (%s, %d, etc.)
	// This is different from GetWithLocale which uses MessageFormat variables ({name})

	// First, let's load a sprintf-style translation
	err = bundle.LoadMessages(map[string]map[string]string{
		"zh-Hans": {
			"greeting": "你好，%s！",
		},
	})
	if err != nil {
		panic(err)
	}

	text, locale, err = localizer.GetfWithLocale("greeting", "Alice")
	printResult("greeting (sprintf)", text, locale, err)

	// Missing key with GetfWithLocale - returns key as-is without formatting
	// to avoid "%!EXTRA" errors from fmt.Sprintf
	text, locale, err = localizer.GetfWithLocale("missing_key", "Alice")
	printResult("missing_key (sprintf)", text, locale, err)
}

func printResult(key, text, locale string, err error) {
	fmt.Printf("Key: %q\n", key)
	fmt.Printf("  Text:   %q\n", text)
	fmt.Printf("  Locale: %q\n", locale)

	if err == nil {
		fmt.Println("  Error:  none")
	} else {
		if errors.Is(err, i18n.ErrTranslationNotFound) {
			fmt.Println("  Error:  ErrTranslationNotFound")
		}
		if errors.Is(err, i18n.ErrMessageFormatCompilation) {
			fmt.Println("  Error:  ErrMessageFormatCompilation")
		}
	}
	fmt.Println()
}

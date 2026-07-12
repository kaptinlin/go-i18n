package main

import (
	"fmt"
	"log"

	"github.com/kaptinlin/go-i18n"
)

func main() {
	bundle, err := i18n.NewBundle("en",
		i18n.WithLocales("zh-Hans"),
	)
	if err != nil {
		panic(err)
	}

	if err := bundle.LoadMessages(map[string]map[string]string{
		"en": {
			"hello":   "Hello, {name}!",
			"goodbye": "Goodbye, {name}!",
		},
		"zh-Hans": {
			// "goodbye" is omitted to demonstrate default-locale fallback.
			"hello": "你好，{name}！",
		},
	}); err != nil {
		panic(err)
	}

	localizer := bundle.NewLocalizer("zh-Hans")

	fmt.Println("=== Lookup Examples ===")

	r := mustLookup(localizer, "hello", i18n.Vars{"name": "World"})
	printResult("hello", &r)

	r = mustLookup(localizer, "goodbye", i18n.Vars{"name": "World"})
	printResult("goodbye", &r)

	r = mustLookup(localizer, "unknown_key")
	printResult("unknown_key", &r)

	fmt.Println("\n=== Detecting Fallback vs Direct Hit ===")

	r = mustLookup(localizer, "hello", i18n.Vars{"name": "World"})
	printSource("hello", &r)

	r = mustLookup(localizer, "goodbye", i18n.Vars{"name": "World"})
	printSource("goodbye", &r)

	fmt.Println("\n=== Context Disambiguation ===")

	if err := bundle.LoadMessages(map[string]map[string]string{
		"zh-Hans": {
			"Post <verb>": "发布文章",
			"Post <noun>": "文章",
		},
	}); err != nil {
		panic(err)
	}

	fmt.Printf("  GetX verb: %s\n", localizer.GetX("Post", "verb"))
	fmt.Printf("  GetX noun: %s\n", localizer.GetX("Post", "noun"))

	r = mustLookup(localizer, "Post <verb>")
	printResult("Post <verb>", &r)
}

func mustLookup(localizer *i18n.Localizer, key string, data ...i18n.Vars) i18n.TranslationResult {
	result, err := localizer.Lookup(key, data...)
	if err != nil {
		log.Fatalf("lookup %q: %v", key, err)
	}
	return result
}

func printResult(key string, r *i18n.TranslationResult) {
	fmt.Printf("Key: %q\n", key)
	fmt.Printf("  Text:           %q\n", r.Text)
	fmt.Printf("  Template:       %q\n", r.Template)
	fmt.Printf("  Matched locale: %q\n", r.MatchedLocale)
	fmt.Printf("  Catalog locale: %q\n", r.CatalogLocale)
	fmt.Printf("  Source:         %q\n", r.Source)
	fmt.Println()
}

func printSource(key string, r *i18n.TranslationResult) {
	switch r.Source {
	case i18n.TranslationSourceMissing:
		fmt.Printf("  %q: NOT FOUND\n", key)
	case i18n.TranslationSourceFallback:
		fmt.Printf("  %q: fallback from %s\n", key, r.CatalogLocale)
	case i18n.TranslationSourceDirect:
		fmt.Printf("  %q: direct hit in %s\n", key, r.CatalogLocale)
	}
}

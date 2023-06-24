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
		"en": map[string]string{
			"message_basic":       "{count, plural, one {Message} other {Messages}}",
			"message_with_number": "{count, plural, =0 {No messages} one {1 message} other {# messages}}",
			"message_with_multiline": `{count, plural,
				=0 {No messages}
				one {1 message}
				other {# messages}
			}`,
			"message_with_select":  "{gender, select, male {He} female {She} other {They}} replied to your message.",
			"message_with_ordinal": "The {floor, selectordinal, one{#st} two{#nd} few{#rd} other{#th}} floor.",
		},
		"zh-Hans": map[string]string{
			"message_basic":       "{count, plural, one {消息} other {消息}}",
			"message_with_number": "{count, plural, =0 {没有消息} one {1 条消息} other {# 条消息}}",
			"message_with_multiline": `{count, plural,
				=0 {没有消息}
				one {1 条消息}
				other {# 条消息}
			}`,
			"message_with_select":  "{gender, select, male {他} female {她} other {他们}}回复了你的消息.",
			"message_with_ordinal": "第 {floor, selectordinal, one{#} two{#} few{#} other{#}} 名.",
		},
	})

	if err != nil {
		fmt.Println(err)
	}

	localizer := bundle.NewLocalizer("en")

	// Output: Message
	fmt.Println(localizer.Get("message_basic", i18n.Vars{
		"count": 1,
	}))

	// Output: Messages
	fmt.Println(localizer.Get("message_basic", i18n.Vars{
		"count": 2,
	}))

	// Output: No messages
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 0,
	}))

	// Output: 1 message
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 1,
	}))

	// Output: 2 messages
	fmt.Println(localizer.Get("message_with_number", i18n.Vars{
		"count": 2,
	}))

	// Output: No messages
	fmt.Println(localizer.Get("message_with_multiline", i18n.Vars{
		"count": 0,
	}))

	// Output: 1 message
	fmt.Println(localizer.Get("message_with_multiline", i18n.Vars{
		"count": 1,
	}))

	// Output: 2 messages
	fmt.Println(localizer.Get("message_with_multiline", i18n.Vars{
		"count": 2,
	}))

	// Output: He replied to your message
	fmt.Println(localizer.Get("message_with_select", i18n.Vars{
		"gender": "male",
	}))

	// Output: She replied to your message
	fmt.Println(localizer.Get("message_with_select", i18n.Vars{
		"gender": "female",
	}))

	// Output: They replied to your message
	fmt.Println(localizer.Get("message_with_select", i18n.Vars{
		"gender": "other",
	}))

	// Output: 1st message
	fmt.Println(localizer.Get("message_with_ordinal", i18n.Vars{
		"floor": 1,
	}))

	// Output: 2nd message
	fmt.Println(localizer.Get("message_with_ordinal", i18n.Vars{
		"floor": 2,
	}))

	// Output: 3rd message
	fmt.Println(localizer.Get("message_with_ordinal", i18n.Vars{
		"floor": 3,
	}))

	// Output: 4th message
	fmt.Println(localizer.Get("message_with_ordinal", i18n.Vars{
		"floor": 4,
	}))
}

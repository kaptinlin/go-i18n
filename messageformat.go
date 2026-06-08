package i18n

import (
	"errors"
	"fmt"
	"maps"

	mf "github.com/kaptinlin/messageformat-go/mf1"
	"golang.org/x/text/language"
)

// ErrMessageFormatCompilation indicates that MessageFormat template compilation failed.
// The translation text is returned as-is without formatting capabilities.
var ErrMessageFormatCompilation = errors.New("messageformat compilation failed")

type messageFunction func(any) (any, error)

type messageFormatter struct {
	options *mf.MessageFormatOptions
}

// WithMessageFormatOptions sets MessageFormat options for the bundle.
func WithMessageFormatOptions(opts *mf.MessageFormatOptions) Option {
	return func(i *I18n) {
		i.messageFormat.setOptions(opts)
	}
}

// WithCustomFormatters adds custom formatters for MessageFormat.
func WithCustomFormatters(formatters map[string]any) Option {
	return func(i *I18n) {
		i.messageFormat.setCustomFormatters(formatters)
	}
}

// WithStrictMode enables strict parsing mode for MessageFormat.
func WithStrictMode(strict bool) Option {
	return func(i *I18n) {
		i.messageFormat.ensureOptions().Strict = strict
	}
}

func (f *messageFormatter) setOptions(opts *mf.MessageFormatOptions) {
	if opts == nil {
		f.options = nil
		return
	}

	f.options = cloneMessageFormatOptions(opts)
}

func (f *messageFormatter) setCustomFormatters(formatters map[string]any) {
	f.ensureOptions().CustomFormatters = maps.Clone(formatters)
}

func (f *messageFormatter) ensureOptions() *mf.MessageFormatOptions {
	if f.options == nil {
		f.options = &mf.MessageFormatOptions{}
	}
	return f.options
}

func cloneMessageFormatOptions(opts *mf.MessageFormatOptions) *mf.MessageFormatOptions {
	cloned := new(mf.MessageFormatOptions)
	*cloned = *opts
	cloned.CustomFormatters = maps.Clone(opts.CustomFormatters)
	return cloned
}

func (f *messageFormatter) compileTranslation(locale, name, text string) (messageFunction, error) {
	formatter, err := f.newFormatter(locale)
	if err != nil {
		return nil, fmt.Errorf("%w for locale %q key %q: %w", ErrMessageFormatCompilation, locale, name, err)
	}

	compiled, err := formatter.Compile(text)
	if err != nil {
		return nil, fmt.Errorf("%w for locale %q key %q: %w", ErrMessageFormatCompilation, locale, name, err)
	}

	return messageFunction(compiled), nil
}

func (f *messageFormatter) format(locale, message string, data []Vars) (string, error) {
	formatter, err := f.newFormatter(locale)
	if err != nil {
		return "", err
	}

	compiled, err := formatter.Compile(message)
	if err != nil {
		return "", fmt.Errorf("%w: compile message: %w", ErrMessageFormatCompilation, err)
	}

	result, err := formatCompiled(messageFunction(compiled), data)
	if err != nil {
		return "", fmt.Errorf("format message: %w", err)
	}
	return result, nil
}

func (f *messageFormatter) newFormatter(locale string) (*mf.MessageFormat, error) {
	base, err := messageFormatBase(locale)
	if err != nil {
		return nil, err
	}

	formatter, err := mf.New(base, f.options)
	if err != nil {
		return nil, fmt.Errorf("create formatter: %w", err)
	}
	return formatter, nil
}

func messageFormatBase(locale string) (string, error) {
	tag, err := language.Parse(locale)
	if err != nil {
		return "", fmt.Errorf("parse locale %q: %w", locale, err)
	}
	base, _ := tag.Base()
	return base.String(), nil
}

func formatCompiled(format messageFunction, data []Vars) (string, error) {
	result, err := format(varsToParams(data))
	if err != nil {
		return "", err
	}
	return stringifyMessageResult(result), nil
}

func stringifyMessageResult(result any) string {
	if str, ok := result.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", result)
}

func varsToParams(vars []Vars) any {
	if len(vars) == 0 {
		return nil
	}
	var params map[string]any
	for _, v := range vars {
		if v == nil {
			continue
		}
		if params == nil {
			params = make(map[string]any, len(v))
		}
		for key, value := range v {
			params[key] = value
		}
	}
	return params
}

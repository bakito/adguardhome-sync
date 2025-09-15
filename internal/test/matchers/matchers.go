package matchers

import (
	"strings"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

type equalIgnoringLineEndingsMatcher struct {
	expected string
}

func (matcher *equalIgnoringLineEndingsMatcher) Match(actual any) (success bool, err error) {
	actualStr, ok := actual.(string)
	if !ok {
		return false, nil
	}

	normalizedActual := strings.ReplaceAll(actualStr, "\r\n", "\n")
	normalizedExpected := strings.ReplaceAll(matcher.expected, "\r\n", "\n")

	return normalizedActual == normalizedExpected, nil
}

func (matcher *equalIgnoringLineEndingsMatcher) FailureMessage(actual any) (message string) {
	actualString, actualOK := actual.(string)
	if actualOK {
		return format.MessageWithDiff(actualString, "to equal", matcher.expected)
	}

	return format.Message(actual, "to equal", matcher.expected)
}

func (matcher *equalIgnoringLineEndingsMatcher) NegatedFailureMessage(actual any) (message string) {
	return format.Message(actual, "not to equal", matcher.expected)
}

// EqualIgnoringLineEndings returns a new matcher.
func EqualIgnoringLineEndings(expected string) types.GomegaMatcher {
	return &equalIgnoringLineEndingsMatcher{
		expected: expected,
	}
}

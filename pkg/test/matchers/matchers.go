package matchers

import (
	"strings"

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
	return "Expected strings to be equal (ignoring line endings)"
}

func (matcher *equalIgnoringLineEndingsMatcher) NegatedFailureMessage(actual any) (message string) {
	return "Expected strings not to be equal (ignoring line endings)"
}

// EqualIgnoringLineEndings returns a new matcher.
func EqualIgnoringLineEndings(expected string) types.GomegaMatcher {
	return &equalIgnoringLineEndingsMatcher{
		expected: expected,
	}
}

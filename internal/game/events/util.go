package events

import (
	"fmt"
	"strings"
)

// Validation utilities

type issueReporter struct {
	issueList  []string
	namePrefix string
}

func newIssueReporter(name ...string) *issueReporter {
	prefix := strings.Join(name, ".")
	return &issueReporter{
		issueList:  []string{},
		namePrefix: prefix,
	}
}

func (i *issueReporter) addf(name string, format string, args ...any) {
	i.issueList = append(i.issueList, fmt.Sprintf("%s: %s", i.namePrefix+name, fmt.Sprintf(format, args...)))
}

func (i *issueReporter) report() error {
	if len(i.issueList) == 0 {
		return nil
	}
	return fmt.Errorf("invalid effect: %v", strings.Join(i.issueList, ", "))
}

func (i *issueReporter) sub(name string) *issueReporter {
	return &issueReporter{
		issueList:  i.issueList,
		namePrefix: i.namePrefix + name,
	}
}

func (i *issueReporter) requireString(name string, value string) *issueReporter {
	if value != "" {
		return i
	}
	i.addf(name, "is required")
	return i
}

func (i *issueReporter) requireCondition(name string, isValid bool) *issueReporter {
	if isValid {
		return i
	}
	i.addf(name, "is invalid")
	return i
}

func (i *issueReporter) requirePositive(name string, value float64) *issueReporter {
	if value > 0 {
		return i
	}
	i.addf(name, "must be positive, got %f", value)
	return i
}

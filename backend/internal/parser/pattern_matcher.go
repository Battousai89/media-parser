package parser

import (
	"regexp"
	"strings"

	"github.com/media-parser/backend/internal/model/entity"
)

type PatternMatcher struct {
	compiledPatterns map[int]*regexp.Regexp
}

func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		compiledPatterns: make(map[int]*regexp.Regexp),
	}
}

func (m *PatternMatcher) Compile(pattern *entity.Pattern) (*regexp.Regexp, error) {
	re, err := regexp.Compile(pattern.Regex)
	if err != nil {
		return nil, err
	}
	m.compiledPatterns[pattern.ID] = re
	return re, nil
}

func (m *PatternMatcher) CompileAll(patterns []*entity.Pattern) error {
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern.Regex)
		if err != nil {
			continue
		}
		m.compiledPatterns[pattern.ID] = re
	}
	return nil
}

func (m *PatternMatcher) Match(html string, pattern *entity.Pattern) []string {
	re, exists := m.compiledPatterns[pattern.ID]
	if !exists {
		var err error
		re, err = m.Compile(pattern)
		if err != nil {
			return nil
		}
	}

	matches := re.FindAllStringSubmatch(html, -1)
	results := make([]string, 0, len(matches))

	for _, match := range matches {
		if len(match) > 1 {
			url := extractURL(match[1])
			if url != "" {
				results = append(results, url)
			}
		} else if len(match) == 1 {
			url := extractURL(match[0])
			if url != "" {
				results = append(results, url)
			}
		}
	}

	return results
}

func (m *PatternMatcher) MatchAll(html string, patterns []*entity.Pattern) map[int][]string {
	results := make(map[int][]string)

	for _, pattern := range patterns {
		matches := m.Match(html, pattern)
		if len(matches) > 0 {
			results[pattern.MediaTypeID] = append(results[pattern.MediaTypeID], matches...)
		}
	}

	return results
}

func extractURL(url string) string {
	url = strings.TrimSpace(url)
	url = strings.Trim(url, `"'`)
	url = strings.TrimSpace(url)

	if strings.HasPrefix(url, "data:") || strings.HasPrefix(url, "javascript:") {
		return ""
	}

	return url
}

func ExtractSrcset(srcset string) []string {
	if srcset == "" {
		return nil
	}

	var urls []string
	parts := strings.Split(srcset, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		fields := strings.Fields(part)
		if len(fields) > 0 {
			url := strings.Trim(fields[0], `"'`)
			if url != "" && !strings.HasPrefix(url, "data:") {
				urls = append(urls, url)
			}
		}
	}

	return urls
}

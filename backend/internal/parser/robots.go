package parser

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RobotsChecker struct {
	cache           map[string]*RobotsRules
	cacheMu         sync.RWMutex
	cacheTTL        time.Duration
	userAgent       string
	ignoreRobotsTxt bool
}

type RobotsRules struct {
	allowed    []string
	disallowed []string
	crawlDelay time.Duration
	fetchedAt  time.Time
}

func NewRobotsChecker(cacheTTL time.Duration, userAgent string, ignoreRobotsTxt bool) *RobotsChecker {
	if cacheTTL == 0 {
		cacheTTL = 24 * time.Hour
	}
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	}

	return &RobotsChecker{
		cache:           make(map[string]*RobotsRules),
		cacheTTL:        cacheTTL,
		userAgent:       userAgent,
		ignoreRobotsTxt: ignoreRobotsTxt,
	}
}

func (c *RobotsChecker) CanFetch(fetchURL string) (bool, error) {
	if c.ignoreRobotsTxt {
		return true, nil
	}

	parsedURL, err := url.Parse(fetchURL)
	if err != nil {
		return false, err
	}

	baseURL := parsedURL.Scheme + "://" + parsedURL.Host
	robotsURL := baseURL + "/robots.txt"

	rules, err := c.getRules(robotsURL)
	if err != nil {
		return true, nil
	}

	path := parsedURL.Path
	if parsedURL.RawQuery != "" {
		path += "?" + parsedURL.RawQuery
	}

	return c.isAllowed(rules, path), nil
}

func (c *RobotsChecker) getRules(robotsURL string) (*RobotsRules, error) {
	c.cacheMu.RLock()
	rules, exists := c.cache[robotsURL]
	c.cacheMu.RUnlock()

	if exists && time.Since(rules.fetchedAt) < c.cacheTTL {
		return rules, nil
	}

	rules, err := c.fetchRules(robotsURL)
	if err != nil {
		return nil, err
	}

	c.cacheMu.Lock()
	c.cache[robotsURL] = rules
	c.cacheMu.Unlock()

	return rules, nil
}

func (c *RobotsChecker) fetchRules(robotsURL string) (*RobotsRules, error) {
	rules := &RobotsRules{
		fetchedAt:  time.Now(),
		allowed:    []string{},
		disallowed: []string{},
	}

	resp, err := http.Get(robotsURL)
	if err != nil {
		return rules, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			c.parseRobotsTxt(string(body), rules)
		}
	}

	return rules, nil
}

func (c *RobotsChecker) parseRobotsTxt(content string, rules *RobotsRules) {
	lines := strings.Split(content, "\n")
	var currentAgent string
	var specificAgentFound bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		directive := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch strings.ToLower(directive) {
		case "user-agent":
			currentAgent = value
			if value == c.userAgent {
				specificAgentFound = true
			}
		case "disallow":
			// Игнорируем wildcard (*) если найден специфичный user-agent
			if currentAgent == c.userAgent || (currentAgent == "*" && !specificAgentFound) {
				rules.disallowed = append(rules.disallowed, value)
			}
		case "allow":
			// Игнорируем wildcard (*) если найден специфичный user-agent
			if currentAgent == c.userAgent || (currentAgent == "*" && !specificAgentFound) {
				rules.allowed = append(rules.allowed, value)
			}
		case "crawl-delay":
			if currentAgent == c.userAgent {
				if delay, err := strconv.Atoi(value); err == nil {
					rules.crawlDelay = time.Duration(delay) * time.Second
				}
			}
		}
	}
}

func (c *RobotsChecker) isAllowed(rules *RobotsRules, path string) bool {
	for _, pattern := range rules.disallowed {
		if matchPattern(pattern, path) {
			return false
		}
	}

	if len(rules.allowed) > 0 {
		for _, pattern := range rules.allowed {
			if matchPattern(pattern, path) {
				return true
			}
		}
		return false
	}

	return true
}

func matchPattern(pattern, path string) bool {
	if pattern == "" {
		return false
	}

	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	return path == pattern
}

func (c *RobotsChecker) ClearCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]*RobotsRules)
}

func (c *RobotsChecker) CleanExpired() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	now := time.Now()
	for u, rules := range c.cache {
		if now.Sub(rules.fetchedAt) > c.cacheTTL {
			delete(c.cache, u)
		}
	}
}

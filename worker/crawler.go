package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"urltracker/internal/cache"

	"github.com/gocolly/colly/v2"
)

type AnalysisResult struct {
	Title             string         `json:"title"`
	HTMLVersion       string         `json:"html_version"`
	HeadingCounts     map[string]int `json:"heading_counts"`
	InternalLinks     int            `json:"internal_links"`
	ExternalLinks     int            `json:"external_links"`
	InaccessibleLinks int            `json:"inaccessible_links"`
	HasLoginForm      bool           `json:"has_login_form"`
	Error             string         `json:"error,omitempty"`
}

func CrawlURL(tracker *cache.URLTracker) (string, error) {
	c := colly.NewCollector()

	result := &AnalysisResult{
		HeadingCounts: make(map[string]int),
	}

	c.OnHTML("html", func(e *colly.HTMLElement) {
		body := strings.ToLower(string(e.Response.Body))

		switch {
		case strings.Contains(body, "<!doctype html>"):
			result.HTMLVersion = "HTML5"
		case strings.Contains(body, "HTML 4.01"):
			result.HTMLVersion = "HTML 4.01"
		default:
			result.HTMLVersion = "Unknown"
		}
	})

	c.OnHTML("head > title", func(e *colly.HTMLElement) {
		result.Title = e.Text
	})

	for i := 1; i <= 6; i++ {
		hLevel := fmt.Sprintf("h%d", i)
		c.OnHTML(hLevel, func(e *colly.HTMLElement) {
			result.HeadingCounts[hLevel]++
		})
	}

	pageURL, urlErr := url.Parse(tracker.URL)
	if urlErr != nil {
		return "", fmt.Errorf("invalid URL: %w", urlErr)
	}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := strings.TrimSpace(e.Attr("href"))

		if href == "" || href == "#" {
			result.InaccessibleLinks++
			return
		}

		linkURL, err := url.Parse(href)
		if err != nil {
			result.InaccessibleLinks++
			return
		}

		if !linkURL.IsAbs() {
			linkURL = pageURL.ResolveReference(linkURL)
		}

		if linkURL.Host == pageURL.Host {
			result.InternalLinks++
		} else {
			result.ExternalLinks++
		}
	})

	c.OnHTML("form", func(e *colly.HTMLElement) {
		e.ForEach("input[type=password], input[name*=pass], input[name*=login]", func(_ int, _ *colly.HTMLElement) {
			result.HasLoginForm = true
		})
	})

	var onError error
	c.OnError(func(e *colly.Response, err error) {
		onError = fmt.Errorf("Failed to collect data for URL. \nError: %w \n Status Code: %d", err, e.StatusCode)
	})

	c.Visit(tracker.URL)
	if onError != nil {
		return "", onError
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

package main

import (
	"strings"
	"time"

	"net/http"
	"net/url"
	"path/filepath"

	"golang.org/x/net/html"
)

type Page struct {
	url   string
	depth int
}

func (page Page) ExtractLinks() []string {
	response, success := connectWebsite(page.url)

	if !success {
		return nil
	}

	links := extractLinks(page.url, response)

	return links
}

func extractLinks(baseUrl string, response *http.Response) []string {
	defer response.Body.Close()

	tokenizer := html.NewTokenizer(response.Body)

	links := make([]string, 0)

	for {
		tokenType := tokenizer.Next()

		if tokenType == html.ErrorToken {
			return links
		}

		token := tokenizer.Token()

		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {

			for _, attr := range token.Attr {
				if attr.Key == "href" {
					link, err := resolveUrl(baseUrl, attr.Val)

					if err == nil && link != "" {
						links = append(links, link)
					}
				}
			}
		}
	}
}

func connectWebsite(url string) (*http.Response, bool) {
	nilResponse := http.Response{}

	client := http.Client{
		Timeout: 60 * time.Second,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &nilResponse, false
	}

	response, err := client.Do(request)

	if err != nil {
		return &nilResponse, false
	}

	return response, true
}

func resolveUrl(baseURL string, path string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	if index := strings.Index(path, "#"); index != -1 {
		path = path[:index]
	}

	if len(path) <= 0 {
		return baseURL, nil
	}

	absolutePath, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	// If the link is already absolute, use it as is
	if absolutePath.IsAbs() {
		if absolutePath.Host == base.Host {
			return absolutePath.String(), nil
		} else {
			return "", nil
		}
	}

	if path[0] == '/' {
		base.Path = "/"
	}

	absolutePath = base.ResolveReference(absolutePath)

	absolutePath.Path = filepath.Clean(absolutePath.Path)

	return absolutePath.String(), nil
}

package main

import (
	"sort"
	"sync"
)

type SafeUrlMap struct {
	v   map[string]bool
	mux sync.Mutex
}

func (c *SafeUrlMap) Visit(key string) {
	c.mux.Lock()
	c.v[key] = true
	c.mux.Unlock()
}

func (c *SafeUrlMap) Exist(key string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	_, ok := c.v[key]
	return ok
}

func Crawl(page Page, fetched *SafeUrlMap, wg *sync.WaitGroup, crawled []string) {
	if page.depth <= 0 {
		return
	}
	if found := fetched.Exist(page.url); found {
		return
	}

	pageLinks := page.ExtractLinks()

	// Run go routine for each linked url
	for _, pageLink := range pageLinks {
		wg.Add(1)

		go func(link string) {
			defer wg.Done()

			Crawl(Page{link, page.depth - 1}, fetched, wg, crawled)
			fetched.Visit(link)
		}(pageLink)
	}
}

func Crawler(rootURL string, maxDepth int) []string {
	fetched := SafeUrlMap{v: make(map[string]bool)}
	var crawled []string

	if maxDepth == 0 {
		return crawled
	}

	var wg sync.WaitGroup

	rootPage := Page{url: rootURL, depth: maxDepth}
	Crawl(rootPage, &fetched, &wg, crawled)
	fetched.Visit(rootURL)

	// Wait for all go routines to finish
	wg.Wait()

	for page := range fetched.v {
		crawled = append(crawled, page)
	}

	sort.Strings(crawled)

	return crawled
}

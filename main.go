package main

// Take a user input in the form of a web address

// Return a list of all assets and web pages

import (
	"fmt"

	"github.com/gocolly/colly"
)

func main() {

	// Instaniate the collector
	c := colly.NewCollector(
		// Limit domains to only these
		colly.AllowedDomains("hackerspaces.org", "wiki.hackerspaces.org"),
	)

	// Callback on every href element
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link on page
		// Only links in Allowed Domains is visited
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Print before making a Request
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit("https://hackerspaces.org/")

}

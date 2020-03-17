package main

// Take a user input in the form of a web address

// Return a list of all assets and web pages

import (
	io "bufio"
	f "fmt"
	"os"
	"regexp"
	sc "strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

func main() {

	domainName := getInput()

	pages, srcs := ScrapeAddress(domainName)

	f.Println("---")
	f.Println("Scrape Complete:")

	printPages(pages)
	f.Println("--------")
	printAssets(srcs)

	writeInfo(pages, srcs, domainName)

}

// ScrapeAddress handles all scraping and associated methods
// Returns two slices one for pages and one for assets
func ScrapeAddress(address string) ([]string, []string) {
	pages, srcs := []string{}, []string{}

	// Instaniate the collector
	c := colly.NewCollector(
		// Limit to only the specific dom
		colly.AllowedDomains(address, "www."+address),

		// Limit depth so that only the page and links on that page are
		// visited in parallel
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// Rate limit to prevent getting barred
	// Parallel limit to control simultaneous requests
	c.Limit(&colly.LimitRule{
		DomainGlob: "*",
		// Nothing seemed to happen here when I increased the number
		Parallelism: 2,
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	// Print before making a Request
	c.OnRequest(func(r *colly.Request) {
		f.Println("Visiting", r.URL.String())
	})

	// Report found errors
	c.OnError(func(r *colly.Response, err error) {
		f.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Get anything with a href
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		c.Visit(e.Request.AbsoluteURL(link))

	})

	// Get anything with a src attribute
	c.OnHTML("[src]", func(e *colly.HTMLElement) {
		src := string(e.Attr("src"))
		// Save asset
		if !inSlice(srcs, src) {
			srcs = append(srcs, src)
		}
	})

	// When a response for a visit attempt is recieved
	c.OnResponse(func(r *colly.Response) {
		f.Println("Visited", r.Request.URL)
		link := string(r.Request.URL.String())
		if !inSlice(pages, link) {
			pages = append(pages, link)
		}
	})

	c.Visit("https://" + address)

	c.Wait()

	return pages, srcs
}

// Print the pages array
func printPages(pages []string) {
	f.Printf("Found Pages: (%v)\n", len(pages))
	for _, v := range pages {
		f.Println(strings.ReplaceAll(v, "https://", ""))
	}
}

func printAssets(assets []string) {
	f.Printf("Assets Found: (%v)\n", len(assets))
	for _, v := range assets {
		f.Println(v)
	}
}

// Search slice for the given parameter
func inSlice(s []string, x string) bool {
	for _, v := range s {
		if v == x {
			return true
		}
	}
	return false
}

// Get user's input as string
func getInput() string {

	// This pattern should match any valid TLD (i.e. .com .co.uk etc...)
	urlPattern := `^[a-zA-Z0-9\.]+(\.\w+)(\.\w\w){0,1}$`

	reader := io.NewReader(os.Stdin)
	f.Println("Please give a domain name:")
	f.Println("\"[example.com]\"")

	// Loop until quit or valid address given
	// (does not check for existence, only matches string pattern)
	for {
		f.Print("")
		text, _ := reader.ReadString('\n')

		text = strings.Replace(text, "\n", "", -1)

		match, err := regexp.Match(urlPattern, []byte(text))

		if match == true {
			return text
		}

		if err != nil {
			f.Println(err)
		}

		f.Printf("Recieved %q, please give a valid domain\n", text)
		f.Println("-----")
	}
}

// Write the scraped data to an output file named after the domain
func writeInfo(pages []string, srcs []string, domain string) {

	f.Println("Writing file...")

	fName := domain + "_log.txt"

	// First check the file exists, if it does, wipe it
	if _, err := os.Stat(fName); err == nil {
		err = os.Remove(fName)
		if err != nil {
			f.Println(err)
			return
		}
	}

	// Now recreate a blank file with the name
	file, err := os.Create(domain + "_log.txt")
	if err != nil {
		f.Println(err)
		file.Close()
		return
	}

	// Write the Pages info
	pageHead := sc.Itoa(len(pages)) + " Pages Found: "
	err = writeLine(pageHead, file)
	if err != nil {
		file.Close()
		return
	}

	for _, v := range pages {
		err = writeLine(v, file)
		if err != nil {
			file.Close()
			return
		}
	}

	// The write the Assets info
	err = writeLine("", file)
	if err != nil {
		file.Close()
		return
	}

	assHead := sc.Itoa(len(srcs)) + " Assests Found: "
	err = writeLine(assHead, file)
	if err != nil {
		file.Close()
		return
	}

	for _, v := range srcs {
		err = writeLine(v, file)
		if err != nil {
			file.Close()
			return
		}
	}

	file.Close()
	f.Println("File Write Successful")
	return

}

// Write individual lines to the given file
func writeLine(text string, file *os.File) error {

	_, err := f.Fprintln(file, text)
	if err != nil {
		f.Println(err)
		return err
	}

	return nil

}

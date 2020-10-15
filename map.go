package sitemap

import (
	"encoding/xml"
	"fmt"
	parser "github.com/vshakirova/html-parser"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func get(inputURL string) []string {
	resp, _ := http.Get(inputURL)
	links := parser.Parser(resp.Body)

	defer resp.Body.Close()

	reqURL := resp.Request.URL
	baseURL := &url.URL{
		Host:   reqURL.Host,
		Scheme: reqURL.Scheme,
	}
	base := baseURL.String()

	return filter(getHrefs(base, links), hasPrefix(base))
}

func getHrefs(base string, links []parser.Link) map[string]empty {
	var tmpStr []string
	res := make(map[string]empty)

	for _, link := range links {
		switch {
		case strings.HasPrefix(link.Href, "/"):
			tmpStr = append(tmpStr, base+link.Href)
		case strings.HasPrefix(link.Href, "http"):
			tmpStr = append(tmpStr, link.Href)
		}
	}

	for i, link := range tmpStr {
		if strings.HasSuffix(link, "/") {
			tmpStr[i] = link[0 : len(link)-1]
		}

		res[tmpStr[i]] = empty{}
	}

	return res
}

func filter(links map[string]empty, hasPrefix func(string) bool) (res []string) {
	for link := range links {
		if hasPrefix(link) {
			res = append(res, link)
		}
	}

	return
}

func hasPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}

type empty struct{}

// Gets list of pages using bfs and creates sitemap using GetXml()
func GetPagesList(link string, maxDepth int) (res []string) {
	used := make(map[string]empty)
	var queue = make(map[string]empty)

	nextQueue := map[string]empty{
		link: {},
	}

	for i := 0; i < maxDepth; i++ {
		queue, nextQueue = nextQueue, make(map[string]empty)

		for link := range queue {
			if _, ok := used[link]; ok {
				continue
			}

			used[link] = empty{}

			for _, tmp := range get(link) {
				nextQueue[tmp] = empty{}
			}
		}
	}

	for k := range used {
		res = append(res, k)
	}

	return
}

type loc struct {
	Value string `xml:"loc"`
}
type urlset struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func GetXML(hrefs []string) error {
	xmlns := "http://www.sitemaps.org/schemas/sitemap/0.9"

	toXML := urlset{
		Xmlns: xmlns,
	}
	for _, link := range hrefs {
		toXML.Urls = append(toXML.Urls, loc{link})
	}

	file, _ := os.Create("sitemap.xml")
	enc := xml.NewEncoder(file)
	_, _ = file.WriteString(xml.Header)
	enc.Indent("", " ")

	if err := enc.Encode(toXML); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}


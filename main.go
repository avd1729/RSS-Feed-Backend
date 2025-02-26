package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type Rss struct {
	XMLName xml.Name   `xml:"rss"`
	Channel RssChannel `xml:"channel"`
}

type RssChannel struct {
	Title string    `xml:"title"`
	Items []RssItem `xml:"item"`
}

type RssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

// List of external RSS sources
var rssSources = []string{
	"https://rss.cnn.com/rss/edition.rss",
	"https://feeds.bbci.co.uk/news/rss.xml",
}

func fetchRss(url string, wg *sync.WaitGroup, ch chan<- []RssItem) {
	defer wg.Done()
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching RSS:", err)
		ch <- nil
		return
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var rss Rss
	xml.Unmarshal(data, &rss)
	ch <- rss.Channel.Items
}

func aggregateRss() []RssItem {
	var wg sync.WaitGroup
	ch := make(chan []RssItem, len(rssSources))

	for _, url := range rssSources {
		wg.Add(1)
		go fetchRss(url, &wg, ch)
	}

	wg.Wait()
	close(ch)

	var allItems []RssItem
	for items := range ch {
		if items != nil {
			allItems = append(allItems, items...)
		}
	}
	return allItems
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	items := aggregateRss()
	xmlData, _ := xml.MarshalIndent(Rss{Channel: RssChannel{Title: "Aggregated RSS", Items: items}}, "", "  ")
	w.Header().Set("Content-Type", "application/xml")
	w.Write(xmlData)
}

func main() {
	http.HandleFunc("/rss", rssHandler)
	fmt.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}

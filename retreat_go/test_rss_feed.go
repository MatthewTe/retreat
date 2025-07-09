package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	items := []Item{}
	for i := 1; i <= 3; i++ {
		items = append(items, Item{
			Title:       fmt.Sprintf("Dummy Item %d", i),
			Link:        fmt.Sprintf("http://localhost:8080/items/%d", i),
			Description: fmt.Sprintf("Description for item %d", i),
			GUID:        fmt.Sprintf("%d", i),
		})
	}

	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:       "Dummy RSS Feed",
			Link:        "http://localhost:8080/rss",
			Description: "A test RSS feed with dummy content",
			Items:       items,
		},
	}

	w.Header().Set("Content-Type", "application/rss+xml")
	xml.NewEncoder(w).Encode(rss)
}

func itemHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/items/"):]
	_, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	html := fmt.Sprintf(`<html>
	<head><title>Dummy Item %s</title></head>
	<body>
		<h1>Dummy Item %s</h1>
		<p>This is dummy HTML content for item %s.</p>
	</body>
</html>`, id, id, id)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func RunDevRSSServer() {
	http.HandleFunc("/rss", rssHandler)
	http.HandleFunc("/items/", itemHandler)

	fmt.Println("Server listening on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

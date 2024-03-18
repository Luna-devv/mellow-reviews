package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var client = &http.Client{Timeout: 10 * time.Second}

func main() {
	http.HandleFunc("/", fetchHandler)
	fmt.Println("Server is running on :8080")

	http.ListenAndServe(":8080", nil)

}

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	path := fmt.Sprintf("https://wumpus.store/bot/%s", id)

	if id == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(path)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		http.Error(w, "Failed to fetch URL", http.StatusInternalServerError)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	res, err := client.Do(req)

	if err != nil {
		http.Error(w, "Failed to fetch URL", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch URL - Status Code: "+res.Status, res.StatusCode)
		return
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		http.Error(w, "Failed to parse HTML", http.StatusInternalServerError)
		return
	}

	reviews := []Review{}

	doc.Find("#reviews .css-1ty38eq").Each(func(i int, review *goquery.Selection) {
		authorName := review.Find("b.css-mcpu91")

		authorHref, _ := review.Find(".css-k008qs>a").Attr("href")
		authorId := strings.Split(authorHref, "/")[2]

		authorProxyAvatar, _ := review.Find("img[aria-label*='Avatar']").Attr("src")
		authorAvatar, _ := url.QueryUnescape(strings.Split(strings.Split(authorProxyAvatar, "=")[1], "&")[0])

		content := review.Find("p.css-542wex")

		helpfulStr := review.Find("p.css-dw5ttn")
		helpful, _ := strconv.Atoi(strings.Split(helpfulStr.Text(), " ")[0])

		rating := "N/A"
		svg, _ := review.Find("svg").Attr("style")

		if strings.Contains(svg, "color:#209b6a") {
			rating = "positive"
		} else {
			rating = "negative"
		}

		author := Author{
			Id:        authorId,
			Username:  strings.Split(authorName.Text(), "@")[1],
			AvatarUrl: authorAvatar,
		}

		reviews = append(reviews, Review{
			Author:  author,
			Content: strings.TrimSpace(content.Text()),
			Helpful: helpful,
			Rating:  rating,
		})

	})

	json, err := json.Marshal(reviews)

	if err != nil {
		http.Error(w, "Failed to generate JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

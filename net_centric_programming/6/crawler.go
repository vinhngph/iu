package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Manga struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Genre  string `json:"genre"`
	Author string `json:"author,omitempty"`
}

type Genre struct {
	Name   string  `json:"genre"`
	Mangas []Manga `json:"manga"`
}

type Webtoons struct {
	Genres []Genre `json:"genres"`
}

var genreLinks = map[string]string{
	"Drama":         "https://www.webtoons.com/en/genres/drama",
	"Fantasy":       "https://www.webtoons.com/en/genres/fantasy",
	"Comedy":        "https://www.webtoons.com/en/genres/comedy",
	"Action":        "https://www.webtoons.com/en/genres/action",
	"Romance":       "https://www.webtoons.com/en/genres/romance",
	"Superhero":     "https://www.webtoons.com/en/genres/super_hero",
	"Horror":        "https://www.webtoons.com/en/genres/horror",
	"SciFi":         "https://www.webtoons.com/en/genres/sf",
	"Slice of Life": "https://www.webtoons.com/en/genres/slice-of-life",
	"Thriller":      "https://www.webtoons.com/en/genres/thriller",
}

func main() {
	var result Webtoons

	client := &http.Client{Timeout: 15 * time.Second}

	for genre, url := range genreLinks {
		fmt.Println("Crawling:", genre)
		time.Sleep(2 * time.Second)

		mangas, err := fetchMangas(client, genre, url)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		if len(mangas) > 0 {
			result.Genres = append(result.Genres, Genre{Name: genre, Mangas: mangas})
		}
	}

	saveToFile(result, "webtoons.json")
	fmt.Println("Done!")
}

func fetchMangas(client *http.Client, genre, url string) ([]Manga, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	doc, _ := html.Parse(strings.NewReader(string(body)))

	var mangas []Manga
	cards := findNodes(doc, "a", "class", "card_item")

	for _, card := range cards {
		m := Manga{Genre: genre}
		for _, attr := range card.Attr {
			if attr.Key == "href" {
				m.URL = attr.Val
				break
			}
		}

		titleNode := findNode(card, "p", "class", "subj")
		if titleNode != nil {
			m.Title = strings.TrimSpace(getText(titleNode))
		}

		authorNode := findNode(card, "p", "class", "author")
		if authorNode != nil {
			m.Author = strings.TrimSpace(getText(authorNode))
		}

		if m.Title != "" && m.URL != "" {
			mangas = append(mangas, m)
			if len(mangas) >= 15 {
				break
			}
		}
	}

	return mangas, nil
}

func findNode(n *html.Node, tag, key, val string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		for _, attr := range n.Attr {
			if attr.Key == key && strings.Contains(attr.Val, val) {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNode(c, tag, key, val); found != nil {
			return found
		}
	}
	return nil
}

func findNodes(n *html.Node, tag, key, val string) []*html.Node {
	var result []*html.Node
	if n.Type == html.ElementNode && n.Data == tag {
		for _, attr := range n.Attr {
			if attr.Key == key && strings.Contains(attr.Val, val) {
				result = append(result, n)
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result = append(result, findNodes(c, tag, key, val)...)
	}
	return result
}

func getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getText(c)
	}
	return text
}

func saveToFile(data Webtoons, filename string) {
	content, _ := json.MarshalIndent(data, "", "\t")
	_ = os.WriteFile(filename, content, 0644)
}

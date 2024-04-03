package main

import (
	"encoding/json"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type CategoriesResponse struct {
	Categories []string `json:"categories"`
}

type SubCategoriesResponse struct {
	Category      string   `json:"category"`
	SubCategories []string `json:"subcategories"`
}

func getForumSubCategories(category string) SubCategoriesResponse {
	list := []string{}

	c := colly.NewCollector()
	c.OnHTML("div.navbar-forum div.list-forum-title a.forum-title", func(e *colly.HTMLElement) {
		if e.Text == category {
			e.DOM.ParentsUntil("~ table").Next().Find("td.alt1Active strong").Each(func(_ int, s *goquery.Selection) {
				list = append(list, s.Text())
			})
		}
	})
	c.Visit("https://www.flashback.org")

	return SubCategoriesResponse{
		Category:      category,
		SubCategories: list,
	}
}

func getForumCategories() CategoriesResponse {
	list := []string{}

	c := colly.NewCollector()
	c.OnHTML("div.list-forum-title a.forum-title", func(e *colly.HTMLElement) {
		list = append(list, e.Text)
	})
	c.Visit("https://www.flashback.org")

	return CategoriesResponse{Categories: list}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts/categories/", func(w http.ResponseWriter, r *http.Request) {
		res := getForumCategories()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("GET /posts/categories/{category}/", func(w http.ResponseWriter, r *http.Request) {
		category := r.PathValue("category")
		res := getForumSubCategories(category)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	http.ListenAndServe("localhost:8090", mux)

}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode"

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

func getForumPosts(category, subcategory string) {
	var subCategoryLink string

	c := colly.NewCollector()
	c.OnHTML("div.navbar-forum div.list-forum-title a.forum-title", func(e *colly.HTMLElement) {
		if e.Text == category {
			e.DOM.ParentsUntil("~ table").Next().Find("td.alt1Active strong").Each(func(_ int, s *goquery.Selection) {
				s.Each(func(_ int, s *goquery.Selection) {
					if s.Text() == subcategory {
						href, exists := s.Parent().Attr("href")
						if exists {
							fmt.Println("Href: ", href)
							subCategoryLink = "https://www.flashback.org" + href
						}
					}
				})
			})
		}
	})
	c.Visit("https://www.flashback.org")

	p := colly.NewCollector()

	p.OnHTML("table.table-striped tbody", func(e *colly.HTMLElement) {
		e.DOM.Find("tr").Each(func(_ int, tr *goquery.Selection) {
			tr.Find("td.td_title").Each(func(_ int, td *goquery.Selection) {
				td.Find("div").First().Each(func(_ int, div *goquery.Selection) {
					a := div.Find("a").First()
					href, exists := a.Attr("href")
					if exists {
						formatedHref := strings.Replace(href, "/", "", -1)
						fmt.Println("Post title: ", a.Text())
						fmt.Println("Post ID: ", formatedHref)
					}
				})
				tr.Find("span.smallfont").Each(func(_ int, span *goquery.Selection) {
					b := span.Find("span").First()
					if b.Text() != "" {
						fmt.Println("Post author: ", b.Text())
					}
				})
			})
			tr.Find("td.td_replies").Each(func(_ int, td *goquery.Selection) {
				a := td.Find("div").First()
				b := td.Find("div").Next()
				aText := a.Text()
				bText := b.Text()

				filterFunc := func(r rune) rune {
					if unicode.IsDigit(r) {
						return r
					}
					return -1
				}

				aText = strings.Map(filterFunc, strings.TrimSpace(aText))
				bText = strings.Map(filterFunc, strings.TrimSpace(bText))

				fmt.Println("Post replies: ", aText)
				fmt.Println("Post views: ", bText)
			})
			tr.Find("td.td_last_post").Each(func(_ int, td *goquery.Selection) {
				td.Find("div").First().Each(func(_ int, div *goquery.Selection) {
					lastPost := strings.TrimSpace(div.Text())
					fmt.Println("Last post: ", lastPost)
				})
				td.Find("div").Next().Each(func(_ int, div *goquery.Selection) {
					lastPostAuthor := strings.TrimSpace(div.Text())
					fmt.Println("Last post author: ", lastPostAuthor)
				})
			})
		})
	})

	p.Visit(subCategoryLink)
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

var replicaId string

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	replicaId = hostname
}

func main() {
	mux := http.NewServeMux()

	middleware := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Origin-Replica", replicaId)
			handler(w, r)
		}
	}

	mux.HandleFunc("/posts/categories", middleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res := getForumCategories()
		json.NewEncoder(w).Encode(res)
	}))

	mux.HandleFunc("/posts/categories/{category}", middleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		category := r.PathValue("category")
		res := getForumSubCategories(category)
		json.NewEncoder(w).Encode(res)
	}))

	mux.HandleFunc("/posts/categories/{category}/{subcategory}", middleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		category := r.PathValue("category")
		subcategory := r.PathValue("subcategory")
		getForumPosts(category, subcategory)
	}))

	http.ListenAndServe("localhost:8090", mux)

}

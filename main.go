package main

import (
	"log"
	"net/http"

	"webswap/frontend"
)

var categories = []string{
	"all",
	"funny",
	"notfunny",
}

func main() {
	http.HandleFunc("/", RootHandler)
	http.ListenAndServe(":8787", nil)
}

func categoryValid(category string) bool {
	for i := range categories {
		if categories[i] == category {
			return true
		}
	}
	return false
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	pathData := frontend.ParsePath(path)
	if pathData.Category == "" {
		pathData.Category = categories[0]
	}

	var offering string
	if !categoryValid(pathData.Category) {
		log.Printf("Invalid category '%s' hit", pathData.Category)
		w.WriteHeader(404)
		offering = frontend.PageError("invalid category")
	} else {
		offering = pathData.Category
	}

	pd := frontend.NewPageData(offering, "one", "two")
	if err := frontend.Output(w, pd); err != nil {
		log.Println(err)
	}
}

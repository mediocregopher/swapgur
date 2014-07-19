package main

import (
	"log"
	"net/http"

	"webswap/frontend"
	//"webswap/backend"
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
	status, receiving := bidnessLogic(r)
	w.WriteHeader(status)

	pd := frontend.NewPageData(receiving, categories...)
	if err := frontend.Output(w, pd); err != nil {
		log.Println(err)
	}
}

func bidnessLogic(r *http.Request) (int, string) {
	path := r.URL.Path
	pathData := frontend.ParsePath(path)
	if pathData.Category == "" {
		pathData.Category = categories[0]
	}

	if !categoryValid(pathData.Category) {
		log.Printf("Invalid category '%s' hit", pathData.Category)
		return 404, frontend.PageError("invalid category")

	}
	//offering := r.PostFormValue("offering")

	return 200, pathData.Category
}

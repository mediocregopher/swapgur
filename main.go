package main

import (
	"log"
	"net/http"
	"regexp"

	"swapgur/frontend"
	"swapgur/backend"
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

var welcome = `The rules are easy - give an image, receive an image from a
random person in return. You must use the raw image link (ends in jpg, jpeg,
png, or gif). An example link would be http://i.imgur.com/vHWOYAU.gif.`

var imgurDirectRegex = regexp.MustCompile(`^https?://i\.imgur\.com/[a-zA-Z0-9]+\.(jpg|jpeg|png|gif)$`)

func bidnessLogic(r *http.Request) (int, string) {
	path := r.URL.Path
	pathData := frontend.ParsePath(path)
	if pathData.Category == "" {
		pathData.Category = categories[0]
	}

	if !categoryValid(pathData.Category) {
		log.Printf("Invalid category '%s' hit", pathData.Category)
		return 404, frontend.PageError("Invalid category")

	}

	offering := r.PostFormValue("offering")
	if offering == "" {
		return 200, frontend.PageParagraph(welcome)
	} else if !imgurDirectRegex.MatchString(offering) {
		return 400, frontend.PageError("Invalid URL")
	}

	receiving := backend.Swap(pathData.Category, offering)
	if receiving == "" {
		return 500, frontend.PageError("Internal Server Error :(")
	}

	return 200, frontend.PageImage(receiving)
}

package frontend

import (
	"html/template"
	"io"
	"log"
	"strings"
)

type PageData struct {
	Categories []string
	Receiving  template.HTML
}

func NewPageData(receiving string, categories ...string) *PageData {
	return &PageData{
		Categories: categories,
		Receiving:  template.HTML(receiving),
	}
}

var tpl *template.Template
func init() {
	var err error
	tpl, err = template.ParseFiles("static/index.html")
	if err != nil {
		log.Fatal(err)
	}
}

func PageError(err string) string {
	return `<a class="error">` + err + `</a>`
}

func PageParagraph(p string) string {
	return `<p>` + p + `</p>`
}

func PageImage(url string) string {
	return `<img src="` + url + `" alt="" />`
}

func Output(w io.Writer, pd *PageData) error {
	return tpl.Execute(w, pd)
}

type PathData struct {
	Category string
}

func ParsePath(path string) *PathData {
	pathSplit := strings.Split(path, "/")
	pathData := PathData{}
	if len(pathSplit) > 1 {
		pathData.Category = pathSplit[1]
	}
	return &pathData
}

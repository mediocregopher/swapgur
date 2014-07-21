package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/base64"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	_ "image/gif"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"

	"swapgur/config"
	"swapgur/backend"
	"swapgur/frontend"
)

var categories = []string{
	"random",
	"art",
	"aww",
	"funny",
	"gifs",
	"nature",
	"nsfw",
	"sports",
}

var categoryDefaults = []string{
	"http://i.imgur.com/vHWOYAU.gif", //random
	"http://i.imgur.com/D8CoAEd.jpg", //art
	"http://i.imgur.com/h2EiFHA.jpg", //aww
	"http://i.imgur.com/QnTc8BW.gif", //funny
	"http://i.imgur.com/QnTc8BW.gif", //gifs
	"http://i.imgur.com/Yavzdox.jpg", //nature
	"http://i.imgur.com/N8lL8H6.jpg", //nsfw
	"http://i.imgur.com/gUgkpTx.jpg", //sports
}

var allowedImageTypes = map[string]bool {
	"jpeg": true,
	"jpg": true,
	"png": true,
	"gif": true,
}

func main() {

	for i := range categoryDefaults {
		backend.Swap(categories[i], categoryDefaults[i])
	}

	http.HandleFunc("/", RootHandler)
	log.Println("Starting ListenAndServer")
	http.ListenAndServe(config.ListenAddr, nil)
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
	status, receiving, centered := bidnessLogic(r)
	w.WriteHeader(status)

	pd := frontend.NewPageData(receiving, categories...)
	pd.ReceivingCenter = centered
	if err := frontend.Output(w, pd); err != nil {
		log.Println(err)
	}
}

var internalErr = frontend.PageError("Internal Server Error :(")
var welcome = `The rules are easy - give an image, receive an image from a
random person in return. You must use the raw image link from imgur (e.g.
http://i.imgur.com/vHWOYAU.gif), or upload an image.<br/><br/>Accepted image
types are jpg, png, and gif.`

var imgurDirectRegex = regexp.MustCompile(`^https?://i\.imgur\.com/[a-zA-Z0-9]+\.(jpg|jpeg|png|gif)$`)

func bidnessLogic(r *http.Request) (int, string, bool) {
	pathData := getPathData(r)
	if pathData.category == "" {
		pathData.category = categories[0]
	}

	if !categoryValid(pathData.category) {
		log.Printf("Invalid category '%s' hit", pathData.category)
		return 404, frontend.PageError("Invalid category"), true

	}

	ip, err := determineIP(r)
	if err != nil {
		log.Printf("%s - determing ip")
		return 500, internalErr, true
	}

	offeringLink, offeringFile, err := parseForm(r)
	if err != nil {
		log.Printf("%s - parsing file", err)
		return 500, internalErr, true
	}

	var offering, receiving string
	if !pathData.mooch {
		if offeringLink != "" {
			if !imgurDirectRegex.MatchString(offeringLink) {
				return 400, frontend.PageError("Invalid URL"), true
			}
			offering = offeringLink
		} else if offeringFile != nil { // TODO this is probably wrong
			if offering, err = encodeImage(offeringFile); err != nil {
				log.Printf("%s - encoding image", err)
				return 400, frontend.PageError("Could not validate image"), true
			}
		} else {
			return 200, frontend.PageParagraph(welcome), false
		}

		offeringMD5 := md5Hex(offering)

		if !backend.IPCanSwap(ip, offeringMD5) {
			return 400, frontend.PageError("You have tried to swap that image too many times! Try a different one."), true
		}

		receiving = backend.Swap(pathData.category, offering)
	} else {
		receiving = backend.Get(pathData.category)
	}

	if receiving == "" {
		return 500, internalErr, true
	}

	log.Printf("ip: %s, category: `%s`", ip, pathData.category)

	return 200, frontend.PageImage(receiving), true
}

func determineIP(r *http.Request) (string, error) {
	if fips, ok := r.Header["X-Forwarded-For"]; ok && len(fips) > 0 {
		ipSplit := strings.Split(fips[0], ",")
		return strings.TrimSpace(ipSplit[0]), nil
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	return ip, err
}

func parseForm(r *http.Request) (string, io.Reader, error) {
	var offeringLink string
	var offeringFile io.Reader
	var err error
	if err = r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		// err != nil when the content-type isn't multipart
		return "", nil, nil
	}
	if r.MultipartForm == nil {
		return "", nil, nil
	}
	if _, ok := r.MultipartForm.Value["offering-link"]; ok {
		offeringLink = r.MultipartForm.Value["offering-link"][0]
	}
	if _, ok := r.MultipartForm.File["offering-file"]; ok {
		offeringHeader := r.MultipartForm.File["offering-file"][0]
		offeringFile, err = offeringHeader.Open()
		if err != nil {
			return "", nil, err
		}
	}
	return offeringLink, offeringFile, nil
}

type pathData struct {
	category string
	mooch    bool
}

func getPathData(r *http.Request) *pathData {
	path := r.URL.Path
	pathSplit := strings.Split(path, "/")
	pathData := pathData{}
	if len(pathSplit) > 1 {
		pathData.category = pathSplit[1]
	}

	if r.FormValue("mooch") != "" {
		pathData.mooch = true
	}

	return &pathData
}

func md5Hex(input string) string {
	h := md5.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func encodeImage(img io.Reader) (string, error) {

	imgTee := bytes.NewBuffer(nil)
	img = io.TeeReader(img, imgTee)

	_, t, err := image.DecodeConfig(img)
	if err != nil {
		return "", err
	}

	if _, ok := allowedImageTypes[t]; !ok {
		return "", errors.New("Unknown image type: "+t)
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteString("data:image/"+t+";base64,")
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	if _, err := io.Copy(enc, imgTee); err != nil {
		log.Println(err)
		return "", err
	}
	if _, err := io.Copy(enc, img); err != nil {
		log.Println(err)
		return "", err
	}
	return buf.String(), nil
}

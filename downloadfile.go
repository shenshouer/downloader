package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	fmt.Println("Downloading file...")

	rawURL := "https://d1ohg4ss876yi2.cloudfront.net/golang-resize-image/big.jpg"

	fileURL, err := url.Parse(rawURL)

	if err != nil {
		panic(err)
	}

	path := fileURL.Path

	segments := strings.Split(path, "/")

	fileName := segments[2] // change the number to accommodate changes to the url.Path position

	file, err := os.Create(fileName)

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer file.Close()

	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := check.Get(rawURL) // add a filter to check redirect

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)

	size, err := io.Copy(file, resp.Body)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%s with %v bytes downloaded", fileName, size)
}
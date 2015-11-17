package main

import (
	"net/http"
//	"fmt"
)

func main() {
	panic(
		http.ListenAndServe(":8080", http.FileServer(http.Dir("/Users/goyoo/workspace/go/src/downloader"))))
}

//func main() {
//	h := http.FileServer(http.Dir("../"))
//	http.ListenAndServe(":1789", ce(h))
//}
//func ce(h http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		fmt.Println(r.URL.Path)
//		http.File()
//		h.ServeHTTP(w, r)
//	})
//}
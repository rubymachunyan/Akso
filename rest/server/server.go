package main

import (
	"Akso/rest"
	"fmt"
	"net/http"
	"os"
)

func main() {
	panic(listenAndServe())

}

func listenAndServe() error {
	routers, err := rest.NewRouter()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fail to create handler...")
		return err
	}
	http.ListenAndServe(rest.DefaultStoreEndPoint, routers)
	return nil
}

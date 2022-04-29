package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/delvatt/forkscount/service"
)

func run(url string) (*bytes.Buffer, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	indentData := &bytes.Buffer{}
	err = json.Indent(indentData, data, "", "  ")
	if err != nil {
		return nil, err
	}

	return indentData, nil
}

func main() {
	service := service.NewService(":9000")

	go func() {
		log.Printf("starting service on %s\n", service.Addr)
		log.Println(service.ListenAndServe())
	}()
	defer func() {
		err := service.Shutdown(context.Background())
		if err == nil {
			log.Println("service successfully shutdown")
		}
	}()

	results, err := run(fmt.Sprintf("http://localhost%s", service.Addr))
	if err != nil {
		log.Println(err)
	}

	fmt.Fprintln(os.Stdout, results)
}

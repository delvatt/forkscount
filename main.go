package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/delvatt/forkscount/service"
)

func run(url string, lastCount int) (*bytes.Buffer, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("n", fmt.Sprintf("%d", lastCount))
	req.URL.RawQuery = q.Encode()

	client := http.DefaultClient
	resp, err := client.Do(req)
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
	lastCount := flag.Int("n", 5, "Number of repository projects to fetch.")
	flag.Parse()

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

	results, err := run(fmt.Sprintf("http://localhost%s", service.Addr), *lastCount)
	if err != nil {
		log.Println(err)
	}

	fmt.Fprintln(os.Stdout, results)
}

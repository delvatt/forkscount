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
	"time"

	"github.com/delvatt/forkscount/service"
)

const defaultServiceAddr = "localhost:9000"

func run(url string, lastCount, timeout int) (*bytes.Buffer, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("n", fmt.Sprintf("%d", lastCount))
	req.URL.RawQuery = q.Encode()

	if timeout > 0 {
		ctx, cancel := context.WithTimeout(req.Context(), time.Duration(timeout*100)*time.Millisecond)
		defer cancel()

		req = req.WithContext(ctx)
	}

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

func init() {

	serviceAddr := os.Getenv("FORKSCOUNT_SERVICE_ADDR")
	if serviceAddr == "" {
		log.Printf("missing %q env var, defaulting to %q\n", "FORKSCOUNT_SERVICE_ADDR", defaultServiceAddr)
		log.Printf("you can configure this by setting %q env var\n", "FORKSCOUNT_SERVICE_ADDR")
		os.Setenv("FORKSCOUNT_SERVICE_ADDR", defaultServiceAddr)
	}
}

func main() {
	lastCount := flag.Int("n", 5, "Number of repository projects to fetch.")
	timeout := flag.Int("t", 0, "time (in the 100 Milliseconds) to wait before the request times out.")
	flag.Parse()

	service := service.NewService(os.Getenv("FORKSCOUNT_SERVICE_ADDR"))

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

	results, err := run(fmt.Sprintf("http://%s", service.Addr), *lastCount, *timeout)
	if err != nil {
		log.Println(err)
	}

	fmt.Fprintln(os.Stdout, results)
}

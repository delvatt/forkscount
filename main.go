package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
const defaultLatCount = 5
const timeoutMultipler = 100

var errHTTPStatusNotOK = errors.New("unexpected http status")

func run(url string, lastCount, timeout int) (*bytes.Buffer, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid http request: %w", err)
	}

	q := req.URL.Query()
	q.Add("n", fmt.Sprintf("%d", lastCount))
	req.URL.RawQuery = q.Encode()

	if timeout > 0 {
		ctx, cancel := context.WithTimeout(req.Context(),
			time.Duration(timeout*timeoutMultipler)*time.Millisecond)
		defer cancel()

		req = req.WithContext(ctx)
	}

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http client error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", errHTTPStatusNotOK, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("invalid response body: %w", err)
	}

	return indentJSONResponse(data)
}

func indentJSONResponse(data []byte) (*bytes.Buffer, error) {
	indentData := &bytes.Buffer{}

	err := json.Indent(indentData, data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("fail to indent JSON response: %w", err)
	}

	return indentData, nil
}

func main() {
	lastCount := flag.Int("n", defaultLatCount, "Number of repository projects to fetch.")
	timeout := flag.Int("t", 0,
		"Time (in the 100 Milliseconds) to wait before the request times out. (default to no timeout)")

	logSuffix := flag.String("l", "", "Suffix name for logging files. (default stdout)")
	flag.Parse()

	var logTarget = os.Stdout

	var err error

	if *logSuffix != "" {
		if os.Getenv("FORKSCOUNT_LOGGING_ENABLED") == "" {
			os.Setenv("FORKSCOUNT_LOGGING_ENABLED", *logSuffix)
		}

		logTarget, err = os.OpenFile(*logSuffix, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		defer logTarget.Close()
	}

	mainLogger := log.New(logTarget, "[MAIN]:", log.LstdFlags)

	var serviceAddr string
	if serviceAddr = os.Getenv("FORKSCOUNT_SERVICE_ADDR"); serviceAddr == "" {
		mainLogger.Printf("missing %q env var, defaulting to %q\n",
			"FORKSCOUNT_SERVICE_ADDR", defaultServiceAddr)

		mainLogger.Printf("you can configure this by setting %q env var\n",
			"FORKSCOUNT_SERVICE_ADDR")

		serviceAddr = defaultServiceAddr
	}

	service := service.NewService(serviceAddr)

	go func() {
		mainLogger.Printf("starting service on %s\n", service.Addr)
		mainLogger.Printf("attempting to stop service: %v", service.ListenAndServe())
	}()

	defer func() {
		err := service.Shutdown(context.Background())
		if err == nil {
			mainLogger.Println("service successfully shutdown")
		}
	}()

	results, err := run(fmt.Sprintf("http://%s", service.Addr), *lastCount, *timeout)
	if err == nil {
		fmt.Fprintln(os.Stdout, results)
	} else {
		mainLogger.Printf("an error occurred while performing the request: %v\n", err)
		fmt.Fprintln(os.Stderr, "error while performing the request. See logs for more details")
	}
}

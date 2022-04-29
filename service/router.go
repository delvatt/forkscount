package service

import "net/http"

func NewService(addr string) *http.Server {
	http.HandleFunc("/", GetLatestProjectJSONHandler)

	return &http.Server{
		Addr:    addr,
		Handler: nil,
	}
}

package main

import "net/http"

type ViaCEPClient interface {
	GetCep(cep string) (*http.Response, error)
}

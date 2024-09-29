package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
	Erro        string `json:"erro"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Informe um cep para saber a temperatura no local")
	})
	http.HandleFunc("/cep", cepHandler)
	fmt.Println("Servidor iniciado na porta 8000")
	http.ListenAndServe(":8000", nil)
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	cep := r.URL.Query().Get("cep")

	//expressão regex para validar se o cep informado possui apenas número e tem 8 dígitos
	regex := regexp.MustCompile(`^\d{8}$`)

	if !regex.MatchString(cep) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprintf(w, "invalid zipcode")
		return
	}

	req, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		w.WriteHeader(req.StatusCode)
		fmt.Fprintf(w, "Erro ao fazer requisição: %s", err)
		return
	}
	defer req.Body.Close()

	res, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Erro ao ler resposta: %s", err)
		return
	}

	var data ViaCEP
	err = json.Unmarshal(res, &data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Erro ao formatar a resposta: %s", err)
		return
	}

	if data.Erro == "true" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "can not find zipcode")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

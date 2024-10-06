package main

import (
	"encoding/json"
	"errors"
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

type WeatherTemperature struct {
	Location struct {
		Name string `json:"name"`
	}
	Current struct {
		TempC float64 `json:"temp_c"`
	}
}

type Temperature struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

var regex *regexp.Regexp

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

	codigo, err := ValidarCep(cep)
	if err != nil {
		w.WriteHeader(codigo)
		fmt.Fprintf(w, err.Error())
		return
	}

	data, codigo, err := GetCep(cep)
	if err != nil {
		w.WriteHeader(codigo)
		fmt.Fprintf(w, err.Error())
	}

	apiKey := "90afc375b7bf4a7cb18171824242909"
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, data.Localidade)

	req, err := http.Get(url)
	if err != nil {
		w.WriteHeader(req.StatusCode)
		fmt.Fprintf(w, "Erro ao fazer requisição da temperatura: %s", err)
		return
	}
	defer req.Body.Close()

	res, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Erro ao ler resposta da temperatura: %s", err)
		return
	}

	var dataWeather WeatherTemperature
	err = json.Unmarshal(res, &dataWeather)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Erro ao formatar a resposta da temperatura: %s", err)
		return
	}

	dataTemperature := Temperature{}
	dataTemperature.TempC = dataWeather.Current.TempC
	dataTemperature.ConverteCelsiusFarenheit(dataWeather.Current.TempC)
	dataTemperature.ConverteCelsiusKelvin(dataWeather.Current.TempC)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dataTemperature)
}

func (t *Temperature) ConverteCelsiusFarenheit(grauCelsius float64) {
	t.TempF = grauCelsius*1.8 + 32
}

func (t *Temperature) ConverteCelsiusKelvin(grauCelsius float64) {
	t.TempK = grauCelsius + 273
}

func ValidarCep(cep string) (int, error) {

	if regex == nil {
		regex = regexp.MustCompile(`^\d{8}$`)
	}

	if !regex.MatchString(cep) {
		return http.StatusUnprocessableEntity, errors.New("invalid zipcode")
	}
	return 0, nil
}

func GetCep(cep string) (*ViaCEP, int, error) {

	req, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		return nil, req.StatusCode, fmt.Errorf("erro ao fazer requisição do Cep: %s", err)
	}
	defer req.Body.Close()

	res, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("erro ao ler resposta do cep: %s", err)
	}

	var data ViaCEP
	err = json.Unmarshal(res, &data)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("erro ao formatar a resposta: %s", err)
	}

	if data.Erro == "true" {
		return nil, http.StatusNotFound, errors.New("can not find zipcode")
	}

	return &data, http.StatusOK, nil

}

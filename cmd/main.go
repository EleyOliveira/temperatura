package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

type ViaCEP struct {
	Localidade string `json:"localidade"`
	Erro       string `json:"erro"`
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
	TempF string  `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

var regex *regexp.Regexp

func main() {

	// Desabilitar a verificação do certificado SSL
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Informe um cep para saber a temperatura no local, exemplo https://temperatura-l5x7giwwma-uc.a.run.app/cep?cep=11700860")
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
		fmt.Fprintf(w, "ocorreu o erro: %s", err.Error())
		return
	}

	data, codigo, err := GetCep(cep)
	if err != nil {
		w.WriteHeader(codigo)
		fmt.Fprintf(w, "ocorreu o erro: %s", err.Error())
		return
	}

	dataTemperature, codigo, err := GetTemperature(data.Localidade)
	if err != nil {
		w.WriteHeader(codigo)
		fmt.Fprintf(w, "ocorreu o erro: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dataTemperature)
}

func (t *Temperature) ConverteCelsiusFarenheit(grauCelsius float64) {
	t.TempF = fmt.Sprintf("%.1f", grauCelsius*1.8+32)
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
		return nil, http.StatusInternalServerError, fmt.Errorf("erro ao fazer requisição da api de CEP: %s", err)
	}
	defer req.Body.Close()

	res, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("erro ao ler resposta do CEP: %s", err)
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

func GetTemperature(localidade string) (*Temperature, int, error) {

	apiKey := "90afc375b7bf4a7cb18171824242909"
	params := url.Values{}
	params.Add("q", localidade)
	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&%s", apiKey, params.Encode())

	req, err := http.Get(url)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("erro ao fazer requisição da api de temperatura: %s", err)
	}
	defer req.Body.Close()

	res, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("erro ao ler resposta da temperatura: %s", err)
	}

	var dataWeather WeatherTemperature
	err = json.Unmarshal(res, &dataWeather)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("erro ao formatar a resposta da temperatura: %s", err)
	}

	dataTemperature := Temperature{}
	dataTemperature.TempC = dataWeather.Current.TempC
	dataTemperature.ConverteCelsiusFarenheit(dataWeather.Current.TempC)
	dataTemperature.ConverteCelsiusKelvin(dataWeather.Current.TempC)

	return &dataTemperature, http.StatusOK, nil

}

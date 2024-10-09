package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

var testTemperature = Temperature{}

type errorReader struct{}

func TestConverteCelsiusFarenheit(t *testing.T) {
	celsius := 27.6
	expected := "81.7"

	testTemperature.ConverteCelsiusFarenheit(celsius)

	if testTemperature.TempF != expected {
		t.Errorf("Esperado %s mas o resultado foi %s", expected, testTemperature.TempF)
	}

}

func TestConverteCelsiusKelvin(t *testing.T) {
	celsius := 23.8
	expected := 296.8

	testTemperature.ConverteCelsiusKelvin(celsius)

	if testTemperature.TempK != expected {
		t.Errorf("Esperado %f mas o resultado foi %f", expected, testTemperature.TempK)
	}
}

func TestValidarCep(t *testing.T) {

	type cenario struct {
		cep       string
		descricao string
	}

	cenarios := []cenario{
		{"011255", "menos de 8 caracteres"},
		{"01@255", "possui caracteres não numéricos"},
		{"", "cep não informado"},
	}

	expectedCode := http.StatusUnprocessableEntity
	expectedMessage := "invalid zipcode"

	for _, cenario := range cenarios {

		t.Run(cenario.descricao, func(t *testing.T) {
			codigo, err := ValidarCep(cenario.cep)
			if codigo != expectedCode {
				t.Errorf("Esperado codigo http %d mas o codigo retornado foi %d", expectedCode, codigo)
			}

			if expectedMessage != err.Error() {
				t.Errorf("Esperado mensagem de erro %s mas a mensagem retornada foi %s", expectedMessage, err.Error())
			}
		})
	}

	cenarioOk := cenario{cep: "06182110", descricao: "cep correto"}
	expectedCode = 0
	codigo, err := ValidarCep(cenarioOk.cep)

	if codigo != expectedCode {
		t.Errorf("Esperado codigo %d para o cenário %s mas o codigo retornado foi %d", expectedCode,
			cenarioOk.descricao, codigo)
	}

	if err != nil {
		t.Errorf("Esperado nil como retorno do erro para o cenario %s mas foi retornado %s", cenarioOk.descricao,
			err.Error())
	}
}

func TestGetCep(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	cep := "06666999"

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep),
		func(req *http.Request) (*http.Response, error) {
			viaCepResponse := &ViaCEP{
				Localidade: "Localidade Teste",
			}
			resp, _ := httpmock.NewJsonResponse(200, viaCepResponse)

			return resp, nil
		})

	_, codigo, _ := GetCep(cep)

	if codigo != 200 {
		t.Errorf("Esperado status code igual a 200, porém foi retornado %d", codigo)
	}

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep),
		func(req *http.Request) (*http.Response, error) {
			viaCepResponse := &ViaCEP{
				Localidade: "Localidade Teste",
				Erro:       "true",
			}
			resp, _ := httpmock.NewJsonResponse(404, viaCepResponse)

			return resp, nil
		})

	_, codigo, err := GetCep(cep)
	expectedCode := http.StatusNotFound
	expectedMessage := "can not find zipcode"

	if codigo != expectedCode {
		t.Errorf("Esperado status code %d, porém foi retornado %d", expectedCode, codigo)
	}

	if expectedMessage != err.Error() {
		t.Errorf("Esperado mensagem %s, porém foi retornado %s", expectedMessage, err.Error())
	}

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep),
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(500, "")
			return resp, nil
		})

	_, codigo, err = GetCep(cep)
	expectedCode = http.StatusInternalServerError
	expectedMessage = "erro ao formatar a resposta"

	if codigo != expectedCode {
		t.Errorf("Esperado status code %d, porém foi retornado %d", expectedCode, codigo)
	}

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("A mensagem %s não contém o texto %s", err.Error(), expectedMessage)
	}

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep),
		httpmock.NewErrorResponder(fmt.Errorf("simulated network error")))

	_, codigo, err = GetCep(cep)

	expectedCode = 500
	expectedMessage = "erro ao fazer requisição da api de CEP:"

	if expectedCode != codigo {
		t.Errorf("Esperado status code %d, porém foi retornado %d", expectedCode, codigo)
	}

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("A mensagem %s não contém o texto %s", err.Error(), expectedMessage)
	}

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep),
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "")
			resp.Body = io.NopCloser(&errorReader{})
			return resp, nil
		})

	_, codigo, err = GetCep(cep)

	expectedCode = 500
	expectedMessage = "erro ao ler resposta do CEP:"

	if expectedCode != codigo {
		t.Errorf("Esperado status code %d, porém foi retornado %d", expectedCode, codigo)
	}

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("A mensagem %s não contém o texto %s", err.Error(), expectedMessage)
	}

}

func TestGetTemperature(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	localidade := "Local Teste"
	apiKey := "90afc375b7bf4a7cb18171824242909"
	params := url.Values{}
	params.Add("q", localidade)

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&%s", apiKey, params.Encode()),
		func(req *http.Request) (*http.Response, error) {
			weather := WeatherTemperature{
				Location: struct {
					Name string `json:"name"`
				}{
					Name: "São Paulo",
				},
				Current: struct {
					TempC float64 `json:"temp_c"`
				}{
					TempC: 25.5,
				},
			}
			resp, _ := httpmock.NewJsonResponse(200, weather)

			return resp, nil
		},
	)

	_, codigo, _ := GetTemperature(localidade)

	if codigo != 200 {
		t.Errorf("Esperado status code igual a 200, porém foi retornado %d", codigo)
	}

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&%s", apiKey, params.Encode()),
		httpmock.NewErrorResponder(fmt.Errorf("simulated network error")))

	_, codigo, err := GetTemperature(localidade)

	expectedCode := 500
	expectedMessage := "erro ao fazer requisição da api de temperatura:"

	if expectedCode != codigo {
		t.Errorf("Esperado status code %d, porém foi retornado %d", expectedCode, codigo)
	}

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("A mensagem %s não contém o texto %s", err.Error(), expectedMessage)
	}

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&%s", apiKey, params.Encode()),
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "")
			resp.Body = io.NopCloser(&errorReader{})
			return resp, nil
		})

	_, codigo, err = GetTemperature(localidade)

	expectedCode = 500
	expectedMessage = "erro ao ler resposta da temperatura:"

	if expectedCode != codigo {
		t.Errorf("Esperado status code %d, porém foi retornado %d", expectedCode, codigo)
	}

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("A mensagem %s não contém o texto %s", err.Error(), expectedMessage)
	}

	httpmock.RegisterResponder("GET", fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&%s", apiKey, params.Encode()),
		func(req *http.Request) (*http.Response, error) {
			resp, _ := httpmock.NewJsonResponse(500, "")
			return resp, nil
		})

	_, codigo, err = GetTemperature(localidade)
	expectedCode = http.StatusInternalServerError
	expectedMessage = "erro ao formatar a resposta"

	if codigo != expectedCode {
		t.Errorf("Esperado status code %d, porém foi retornado %d", expectedCode, codigo)
	}

	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("A mensagem %s não contém o texto %s", err.Error(), expectedMessage)
	}
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

package main

import (
	"net/http"
	"testing"
)

var testTemperature = Temperature{}

func TestConverteCelsiusFarenheit(t *testing.T) {
	celsius := 27.6
	expected := 81.68

	testTemperature.ConverteCelsiusFarenheit(celsius)

	if testTemperature.TempF != expected {
		t.Errorf("Esperado %f mas o resultado foi %f", expected, testTemperature.TempF)
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

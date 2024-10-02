package main

import "testing"

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

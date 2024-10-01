package main

import "testing"

func TestConverteCelsiusFarenheit(t *testing.T) {
	celsius := 27.6
	expected := 81.68

	testTemperature := Temperature{}
	testTemperature.ConverteCelsiusFarenheit(celsius)

	if testTemperature.TempF != expected {
		t.Errorf("Esperado %f mas o resultado foi %f", expected, testTemperature.TempF)
	}

}

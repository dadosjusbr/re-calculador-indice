package main

import (
	"testing"

	"github.com/matryer/is"
)

func TestCalcCompletenessScore(t *testing.T) {
	data := []struct {
		Desc     string
		Input    Meta
		Expected float64
	}{
		{"Sempre positivo", Meta{
			HaveEnrollment:   true,
			ThereIsACapacity: true,
			HasPosition:      true,
			BaseRevenue:      "DETALHADO",
			OtherRecipes:     "DETALHADO",
			Expenditure:      "DETALHADO",
		}, 1.0},
		{"Sempre negativo", Meta{
			HaveEnrollment:   false,
			ThereIsACapacity: false,
			HasPosition:      false,
			BaseRevenue:      "AUSENCIA",
			OtherRecipes:     "AUSENCIA",
			Expenditure:      "AUSENCIA",
		}, 0.0},
		{"CNJ-2020", Meta{
			HaveEnrollment:   false,
			ThereIsACapacity: false,
			HasPosition:      false,
			BaseRevenue:      "DETALHADO",
			OtherRecipes:     "DETALHADO",
			Expenditure:      "DETALHADO",
		}, 0.5},
	}

	for _, d := range data {
		t.Run(d.Desc, func(t *testing.T) {
			is := is.New(t)
			b := calcCompletenessScore(d.Input)
			is.Equal(b, d.Expected)
		})
	}
}

func TestCalcEasinessScore(t *testing.T) {
	data := []struct {
		Desc     string
		Input    Meta
		Expected float64
	}{
		{"Sempre positivo", Meta{
			NoLoginRequired:   true,
			NoCaptchaRequired: true,
			Access:            "ACESSO_DIRETO",
			ConsistentFormat:  true,
			StrictlyTabular:   true,
		}, 1.0},
		{"Sempre negativo", Meta{
			NoLoginRequired:   false,
			NoCaptchaRequired: false,
			Access:            "NECESSITA_SIMULACAO_USUARIO",
			ConsistentFormat:  false,
			StrictlyTabular:   false,
		}, 0.0},
		{"CNJ-2020", Meta{
			NoLoginRequired:   true,
			NoCaptchaRequired: true,
			Access:            "NECESSITA_SIMULACAO_USUARIO",
			ConsistentFormat:  true,
			StrictlyTabular:   true,
		}, 0.8},
	}

	for _, d := range data {
		t.Run(d.Desc, func(t *testing.T) {
			is := is.New(t)
			b := calcEasinessScore(d.Input)
			is.Equal(b, d.Expected)
		})
	}
}

func TestCalcScore(t *testing.T) {
	data := []struct {
		Desc     string
		Input    Meta
		Expected float64
	}{
		{"Sempre positivo", Meta{
			HaveEnrollment:    true,
			ThereIsACapacity:  true,
			HasPosition:       true,
			BaseRevenue:       "DETALHADO",
			OtherRecipes:      "DETALHADO",
			Expenditure:       "DETALHADO",
			NoLoginRequired:   true,
			NoCaptchaRequired: true,
			Access:            "ACESSO_DIRETO",
			ConsistentFormat:  true,
			StrictlyTabular:   true,
		}, 1.0},
		{"Sempre negativo", Meta{
			HaveEnrollment:    false,
			ThereIsACapacity:  false,
			HasPosition:       false,
			BaseRevenue:       "AUSENCIA",
			OtherRecipes:      "AUSENCIA",
			Expenditure:       "AUSENCIA",
			NoLoginRequired:   false,
			NoCaptchaRequired: false,
			Access:            "NECESSITA_SIMULACAO_USUARIO",
			ConsistentFormat:  false,
			StrictlyTabular:   false,
		}, 0.0},
		{"CNJ-2020", Meta{
			HaveEnrollment:    false,
			ThereIsACapacity:  false,
			HasPosition:       false,
			BaseRevenue:       "DETALHADO",
			OtherRecipes:      "DETALHADO",
			Expenditure:       "DETALHADO",
			NoLoginRequired:   true,
			NoCaptchaRequired: true,
			Access:            "NECESSITA_SIMULACAO_USUARIO",
			ConsistentFormat:  true,
			StrictlyTabular:   true,
		}, 0.65},
	}

	for _, d := range data {
		t.Run(d.Desc, func(t *testing.T) {
			is := is.New(t)
			b := calcScore(d.Input)
			is.Equal(b, d.Expected)
		})
	}
}

package country

import (
	"net/http"
	"testing"

	"github.com/ingka-group-digital/ocp-go-utils/api/test"
)

func TestIntegrationCountryHandler_GetCountries(t *testing.T) {
	if testing.Short() {
		t.Skip("(skipped)")
	}

	it := test.NewIntegrationTest(t, test.IntegrationTestWithPostgres{})
	defer func() {
		it.TearDown()
	}()

	countryHandler := NewCountryHandler()

	tests := []test.Data{
		{
			Name:           "ok: Get countries",
			Method:         http.MethodGet,
			Handler:        countryHandler.GetCountries,
			ExpectResponse: "get-countries-ok",
			ExpectCode:     http.StatusOK,
		},
	}

	test.AssertAll(it, tests)
}

package country

import "sort"

// Countries represents the list of countries returned by the handler.
type Countries struct {
	Data []Country `json:"data"`
} // @name Countries

// Country represents a single country in the system.
type Country struct {
	Label string `json:"label"`
	Code  string `json:"code"`
} // @name Country

// SortedCountries returns the default countries sorted.
func SortedCountries() Countries {
	countries := CountriesList

	// Sort by code, keeping original order or equal elements
	sort.SliceStable(countries, func(i, j int) bool {
		return countries[i].Code < countries[j].Code
	})
	return Countries{
		Data: countries,
	}
}

// CountriesList contains the list of countries.
var CountriesList = []Country{
	{
		Label: "Portugal",
		Code:  "PT",
	},
	{
		Label: "Italy",
		Code:  "IT",
	},
	{
		Label: "France",
		Code:  "FR",
	},
	{
		Label: "Norway",
		Code:  "NO",
	},
	{
		Label: "Finland",
		Code:  "FI",
	},
	{
		Label: "Denmark",
		Code:  "DK",
	},
	{
		Label: "Romania",
		Code:  "RO",
	},
	{
		Label: "Slovenia",
		Code:  "SI",
	},
	{
		Label: "Serbia",
		Code:  "RS",
	},
	{
		Label: "Croatia",
		Code:  "HR",
	},
	{
		Label: "Poland",
		Code:  "PL",
	},
	{
		Label: "Slovakia",
		Code:  "SK",
	},
	{
		Label: "Hungary",
		Code:  "HU",
	},
	{
		Label: "Czech Republic",
		Code:  "CZ",
	},
	{
		Label: "Ireland",
		Code:  "IE",
	},
	{
		Label: "Great Britain",
		Code:  "GB",
	},
	{
		Label: "Switzerland",
		Code:  "CH",
	},
	{
		Label: "Belgium",
		Code:  "BE",
	},
	{
		Label: "Austria",
		Code:  "AT",
	},
	{
		Label: "India",
		Code:  "IN",
	},
	{
		Label: "South Korea",
		Code:  "KR",
	},
	{
		Label: "Japan",
		Code:  "JP",
	},
	{
		Label: "China",
		Code:  "CN",
	},
	{
		Label: "Australia",
		Code:  "AU",
	},
	{
		Label: "Netherlands",
		Code:  "NL",
	},
	{
		Label: "Sweden",
		Code:  "SE",
	},
	{
		Label: "Germany",
		Code:  "DE",
	},
	{
		Label: "Spain",
		Code:  "ES",
	},
	{
		Label: "Canada",
		Code:  "CA",
	},
	{
		Label: "United States",
		Code:  "US",
	},
}

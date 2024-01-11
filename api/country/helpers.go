package country

// Exists returns whether a country exists in the countries list.
func Exists(c string) bool {
	for _, country := range CountriesList {
		if country.Code == c {
			return true
		}
	}
	return false
}

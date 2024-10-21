package date

// IKEAFinancialYear returns the financial year for the given year and month
func IKEAFinancialYear(y, m int) int {
	if m >= 9 {
		return y + 1
	}
	return y
}

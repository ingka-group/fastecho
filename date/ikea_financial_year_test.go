package date

import "testing"

func TestIKEAFinancialYear(t *testing.T) {
	tests := []struct {
		name string
		y    int
		m    int
		want int
	}{
		{
			name: "ok: financial year different iso year",
			y:    2024,
			m:    9,
			want: 2025,
		},
		{
			name: "ok: financial year same iso year",
			y:    2024,
			m:    8,
			want: 2024,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IKEAFinancialYear(tt.y, tt.m); got != tt.want {
				t.Errorf("IKEAFinancialYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

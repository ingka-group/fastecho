package country

import "testing"

func TestExists(t *testing.T) {
	tests := []struct {
		name    string
		country string
		want    bool
	}{
		{
			name:    "ok: country exists",
			country: "PT",
			want:    true,
		},
		{
			name:    "fail: country does not exist",
			country: "XX",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exists(tt.country); got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

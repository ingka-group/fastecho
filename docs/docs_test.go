package main

import (
	"testing"

	"github.com/ingka-group-digital/ocp-go-utils/api/test"
)

func Test_generateDocs(t *testing.T) {
	type args struct {
		functions string
		pkg       string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ok: generate docs",
			args: args{
				functions: "fastecho/health.Live",
				pkg:       "fastecho",
			},
			want: "valid-docs-live.txt",
		},
		{
			name: "ok: generate docs",
			args: args{
				functions: "fastecho/health.Live,fastecho/health.Ready",
				pkg:       "fastecho",
			},
			want: "valid-docs-health.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateDocs(tt.args.functions, tt.args.pkg)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateDocs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var fixtures test.Fixtures
			want := fixtures.ReadFixture(tt.want, "")

			if got != want {
				t.Errorf("generateDocs() got\n%v \nwant\n%v", got, want)
			}
		})
	}
}

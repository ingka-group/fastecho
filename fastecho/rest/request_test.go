package rest

import (
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHTTPRequest_SetRequestBodyJSON(t *testing.T) {
	type args struct {
		p *Params
	}

	tests := []struct {
		name string
		args args
		want *HTTPRequest
	}{
		{
			name: "ok: set body",
			args: args{
				p: &Params{
					RequestBody: 10,
				},
			},
		},
		{
			name: "ok: empty body",
			args: args{
				p: &Params{},
			},
			want: &HTTPRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRequest(tt.args.p)
			if err != nil {
				t.Fail()
			}

			err = r.SetRequestBodyJSON(tt.args.p)
			if err != nil {
				t.Fail()
			}

			if tt.args.p.RequestBody != nil {
				assert.True(t, r.Body != nil)
			} else {
				assert.True(t, r.Body == nil)
			}
		})
	}
}

func TestHTTPRequest_SetRequestHeaders(t *testing.T) {
	requestUUID := "3b9b54e9-971e-4da3-a2f0-d70707d895e3"

	type args struct {
		p *Params
	}

	tests := []struct {
		name string
		args args
		want *HTTPRequest
	}{
		{
			name: "ok: set headers with request ID",
			args: args{
				p: &Params{
					RequestHeaders: map[string]string{
						echo.HeaderConnection:    "keep-alive",
						echo.HeaderContentLength: "20",
					},
					RequestID: &requestUUID,
				},
			},
			want: &HTTPRequest{
				Header: map[string][]string{
					echo.HeaderConnection:    {"keep-alive"},
					echo.HeaderContentLength: {"20"},
					echo.HeaderXRequestID:    {requestUUID},
				},
			},
		},
		{
			name: "ok: set headers without request ID",
			args: args{
				p: &Params{
					RequestHeaders: map[string]string{
						echo.HeaderConnection:    "keep-alive",
						echo.HeaderContentLength: "20",
					},
				},
			},
			want: &HTTPRequest{
				Header: map[string][]string{
					echo.HeaderConnection:    {"keep-alive"},
					echo.HeaderContentLength: {"20"},
				},
			},
		},
		{
			name: "ok: empty headers",
			args: args{
				p: &Params{},
			},
			want: &HTTPRequest{
				Header: map[string][]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRequest(tt.args.p)
			if err != nil {
				t.Fail()
			}

			r.SetRequestHeaders(tt.args.p)
			assert.Equal(t, tt.want.Header, r.Header)
		})
	}
}

func TestHTTPRequest_SetQueryParams(t *testing.T) {
	type args struct {
		p *Params
	}

	tests := []struct {
		name string
		args args
		want *HTTPRequest
	}{
		{
			name: "ok: set query parameters",
			args: args{
				p: &Params{
					RequestQueryParams: []QueryParam{
						{
							Key:   "code",
							Value: "2",
						},
						{
							Key:   "country",
							Value: "NL",
						},
					},
				},
			},
			want: &HTTPRequest{
				URL: &url.URL{
					RawQuery: "code=2&country=NL",
				},
			},
		},
		{
			name: "ok: empty query parameters",
			args: args{
				p: &Params{},
			},
			want: &HTTPRequest{
				URL: &url.URL{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRequest(tt.args.p)
			if err != nil {
				t.Fail()
			}

			r.SetQueryParams(tt.args.p)
			assert.Equal(t, tt.want.URL.RawQuery, r.URL.RawQuery)
		})
	}
}

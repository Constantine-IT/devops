package handlers

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Constantine-IT/devops/cmd/server/internal/storage"
)

func TestHandlersResponse(t *testing.T) {

	type want struct {
		statusCode  int
		contentType string
		body        string
	}

	tests := []struct {
		name        string
		request     string
		requestType string
		body        string
		want        want
	}{
		{
			name:        "Test #1: Request with method that not allowed",
			request:     "/update/Alloc/gauge/100",
			requestType: http.MethodPatch,
			body:        "",
			want: want{
				statusCode:  http.StatusMethodNotAllowed,
				contentType: "",
				body:        "",
			},
		},
		{
			name:        "Test #2: Request with too long PATH",
			request:     "/update/1223/232323/232323/32323/23232",
			requestType: http.MethodPost,
			body:        "",
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				body:        "404 page not found\n",
			},
		},
		{
			name:        "Test #3: Request with too short PATH",
			request:     "/update/212323123/gauge",
			requestType: http.MethodPost,
			body:        "",
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
				body:        "404 page not found\n",
			},
		},
		{
			name:        "Test #4: Request with unsupported metrica TYPE",
			request:     "/update/unknown/Alloc/123",
			requestType: http.MethodPost,
			body:        "",
			want: want{
				statusCode:  http.StatusNotImplemented,
				contentType: "text/plain; charset=utf-8",
				body:        "only GAUGE or COUNTER metrica types are allowed\n",
			},
		},
		{
			name:        "Test #5: Request with unsupported metrica VALUE",
			request:     "/update/gauge/Alloc/none",
			requestType: http.MethodPost,
			body:        "",
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				body:        "only GAUGE or COUNTER metrica values are allowed\n",
			},
		},
	}

	app := &Application{
		ErrorLog:   log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		InfoLog:    log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		BaseURL:    "http://127.0.0.1:8080",
		Datasource: &storage.Storage{Data: make(map[string]storage.MetricaRow)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := app.Routes()
			ts := httptest.NewServer(r)
			defer ts.Close()

			resp, body := testSimpleRequest(t, ts, tt.requestType, tt.request, tt.body)
			defer resp.Body.Close()
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-Type"))
			if resp.Header.Get("Content-Type") == "application/json" {
				assert.JSONEq(t, tt.want.body, body)
			} else {
				assert.Equal(t, tt.want.body, body)
			}
		})
	}
}

func testSimpleRequest(t *testing.T, ts *httptest.Server, method, path string, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

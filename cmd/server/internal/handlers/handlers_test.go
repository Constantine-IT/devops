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
				body:        "only GAUGE or COUNTER metrica TYPES are allowed\n",
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
				body:        "only GAUGE or COUNTER metrica VALUES are allowed\n",
			},
		},
		{
			name:        "Test #6: Request with supported metrica - SUCCESS",
			request:     "/update/gauge/Alloc/1000",
			requestType: http.MethodPost,
			body:        "",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "",
				body:        "",
			},
		},
		{
			name:        "Test #7: Request GET metrica value, that previously stored ",
			request:     "/value/gauge/Alloc",
			requestType: http.MethodGet,
			body:        "",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				body:        "1000",
			},
		},
		{
			name:        "Test #8: Request GET all metrica with values",
			request:     "/",
			requestType: http.MethodGet,
			body:        "",
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/html",
				body:        "Metrica: Alloc = 1000\n",
			},
		},
		{
			name:        "Test #9: Request POST to get one metrica value by API",
			request:     "/value",
			requestType: http.MethodPost,
			body:        `{"id":"Alloc", "type":"gauge"}`,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"Alloc", "type":"gauge", "value":1000}`,
			},
		},
		{
			name:        "Test #10: Request POST to update one metrica value by API",
			request:     "/update",
			requestType: http.MethodPost,
			body:        `{"id":"Alloc", "type":"gauge",  "value":0}`,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "",
				body:        ``,
			},
		},
		{
			name:        "Test #11: Request POST to insert one metrica value by API like autotest_4",
			request:     "/update/",
			requestType: http.MethodPost,
			body:        `{"id":"HeapReleased", "type":"gauge", "value":2695168.000000}`,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "",
				body:        ``,
			},
		},
		{
			name:        "Test #12: Request POST to get one metrica value by API like autotest_4",
			request:     "/value/",
			requestType: http.MethodPost,
			body:        `{"id":"HeapReleased", "type":"gauge"}`,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"HeapReleased", "type":"gauge", "value":2695168.000000}`,
			},
		},
		{
			name:        "Test #13: Request POST to put counter value by API",
			request:     "/update/",
			requestType: http.MethodPost,
			body:        `{"id":"HeapCount", "type":"counter", "delta":111111}`,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "",
				body:        ``,
			},
		},
		{
			name:        "Test #14: Request POST to get counter value by API",
			request:     "/value/",
			requestType: http.MethodPost,
			body:        `{"id":"HeapCount", "type":"counter"}`,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				body:        `{"id":"HeapCount", "type":"counter", "delta":111111}`,
			},
		},
		{
			name:        "Test #15: Request POST to put counter value by API in SECOND time",
			request:     "/update/",
			requestType: http.MethodPost,
			body:        `{"id":"HeapCount", "type":"counter", "delta":111111}`,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "",
				body:        ``,
			},
		},
		{
			name:        "Test #16: Request POST to get counter (that was SECOND updated) value by OLD SCHOOL",
			request:     "/value/counter/HeapCount",
			requestType: http.MethodGet,
			body:        ``,
			want: want{
				statusCode:  http.StatusOK,
				contentType: "text/plain; charset=utf-8",
				body:        `222222`,
			},
		},
	}

	app := &Application{
		ErrorLog:   log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		InfoLog:    log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		Datasource: &storage.Storage{Data: make([]storage.Metrics, 0)},
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

package provider

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type TestHttpMock struct {
	server *httptest.Server
}

const testDataSourceConfig_basic = `
data "http" "http_test" {
  url = "%s/meta_%d.txt"
}

output "body" {
  value = data.http.http_test.body
}

output "response_headers" {
  value = data.http.http_test.response_headers
}
`
const testDataSourceConfig_post = `
data "http" "http_test" {
  url = "%s/meta_%d.txt"
  request_method = "POST"
  request_body = "mytest"
}

output "body" {
  value = data.http.http_test.body
}

output "response_headers" {
  value = data.http.http_test.response_headers
}
`

func TestDataSource_http200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_basic, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["data.http.http_test"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["body"].Value != "1.0.0,GET" {
						return fmt.Errorf(
							`'body' output is %s; want '1.0.0,GET'`,
							outputs["body"].Value,
						)
					}

					response_headers := outputs["response_headers"].Value.(map[string]interface{})

					if response_headers["X-Single"].(string) != "foobar" {
						return fmt.Errorf(
							`'X-Single' response header is %s; want 'foobar'`,
							response_headers["X-Single"].(string),
						)
					}

					if response_headers["X-Double"].(string) != "1, 2" {
						return fmt.Errorf(
							`'X-Double' response header is %s; want '1, 2'`,
							response_headers["X-Double"].(string),
						)
					}

					return nil
				},
			},
		},
	})
}

func TestDataSource_post_http200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_post, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["data.http.http_test"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["body"].Value != "1.0.0,POST,mytest" {
						return fmt.Errorf(
							`'body' output is %s; want '1.0.0,POST,mytest'`,
							outputs["body"].Value,
						)
					}

					response_headers := outputs["response_headers"].Value.(map[string]interface{})

					if response_headers["X-Single"].(string) != "foobar" {
						return fmt.Errorf(
							`'X-Single' response header is %s; want 'foobar'`,
							response_headers["X-Single"].(string),
						)
					}

					if response_headers["X-Double"].(string) != "1, 2" {
						return fmt.Errorf(
							`'X-Double' response header is %s; want '1, 2'`,
							response_headers["X-Double"].(string),
						)
					}

					return nil
				},
			},
		},
	})
}

func TestDataSource_http404(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(testDataSourceConfig_basic, testHttpMock.server.URL, 404),
				ExpectError: regexp.MustCompile("HTTP request error. Response code: 404"),
			},
		},
	})
}

const testDataSourceConfig_withHeaders = `
data "http" "http_test" {
  url = "%s/restricted/meta_%d.txt"

  request_headers = {
    "Authorization" = "Zm9vOmJhcg=="
  }
}

output "body" {
  value = data.http.http_test.body
}
`

func TestDataSource_withHeaders200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_withHeaders, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["data.http.http_test"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["body"].Value != "1.0.0" {
						return fmt.Errorf(
							`'body' output is %s; want '1.0.0'`,
							outputs["body"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

const testDataSourceConfig_utf8 = `
data "http" "http_test" {
  url = "%s/utf-8/meta_%d.txt"
}

output "body" {
  value = "${data.http.http_test.body}"
}
`

func TestDataSource_utf8(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_utf8, testHttpMock.server.URL, 200),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["data.http.http_test"]
					if !ok {
						return fmt.Errorf("missing data resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["body"].Value != "1.0.0" {
						return fmt.Errorf(
							`'body' output is %s; want '1.0.0'`,
							outputs["body"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

const testDataSourceConfig_utf16 = `
data "http" "http_test" {
  url = "%s/utf-16/meta_%d.txt"
}

output "body" {
  value = "${data.http.http_test.body}"
}
`

func TestDataSource_utf16(t *testing.T) {
	testHttpMock := setUpMockHttpServer()

	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceConfig_utf16, testHttpMock.server.URL, 200),
				// This should now be a warning, but unsure how to test for it...
				//ExpectWarning: regexp.MustCompile("Content-Type is not a text type. Got: application/json; charset=UTF-16"),
			},
		},
	})
}

func setUpMockHttpServer() *TestHttpMock {
	Server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "text/plain")
			w.Header().Add("X-Single", "foobar")
			w.Header().Add("X-Double", "1")
			w.Header().Add("X-Double", "2")
			if r.URL.Path == "/meta_200.txt" {
				var body bytes.Buffer
				body.WriteString("1.0.0")
				if r.Method == "GET" {
					body.WriteString(",GET")
				} else if r.Method == "POST" {
					buf := new(bytes.Buffer)
					buf.ReadFrom(r.Body)
					newStr := buf.String()
					body.WriteString(fmt.Sprintf(",POST,%s", newStr))
				}
				w.WriteHeader(http.StatusOK)
				w.Write(body.Bytes())
			} else if r.URL.Path == "/restricted/meta_200.txt" {
				if r.Header.Get("Authorization") == "Zm9vOmJhcg==" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("1.0.0"))
				} else {
					w.WriteHeader(http.StatusForbidden)
				}
			} else if r.URL.Path == "/utf-8/meta_200.txt" {
				w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("1.0.0"))
			} else if r.URL.Path == "/utf-16/meta_200.txt" {
				w.Header().Set("Content-Type", "application/json; charset=UTF-16")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("\"1.0.0\""))
			} else if r.URL.Path == "/meta_404.txt" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)

	return &TestHttpMock{
		server: Server,
	}
}

package botta_test

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/starkandwayne/goutils/botta"
)

func expect_body(req *http.Request, content string) {
	body, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	Expect(err).ShouldNot(HaveOccurred())
	Expect(body).Should(Equal([]byte(content)))
}

var _ = Describe("HTTP Helpers", func() {
	Context("HttpRequest()", func() {
		It("should return an http.Request with encoded json if data provided", func() {
			req, err := botta.HttpRequest("GET", "https://localhost:1234/test", map[string]interface{}{"asdf": 1234})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())
			expect_body(req, `{"asdf":1234}`)
		})
		It("should return an http.Request without any payload if no data provided", func() {
			req, err := botta.HttpRequest("GET", "https://localhost:1234/test", nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())
			expect_body(req, "")
		})
		It("should fail if JSON data could not be marshaled", func() {
			req, err := botta.HttpRequest("POST", "https://localhost:1234/test", map[int]string{1: "asdf"})
			Expect(err).Should(HaveOccurred())
			Expect(req).Should(BeNil())
		})
		It("should fail if http.NewRequest failed", func() {
			req, err := botta.HttpRequest("INVALID", "%", nil) // '%' is an invalid URL!
			Expect(req).Should(BeNil())
			Expect(err).Should(HaveOccurred())
		})
		It("should set Content-Type + Accept headers", func() {
			req, err := botta.HttpRequest("GET", "https://localhost:1234/test", nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())

			Expect(req.Header.Get("Content-Type")).Should(Equal("application/json"))
			Expect(req.Header.Get("Accept")).Should(Equal("application/json"))
		})
		It("should use the specified method and URL in the request", func() {
			req, err := botta.HttpRequest("GET", "https://localhost:1234/test", nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())

			Expect(req.Method).Should(Equal("GET"))
			Expect(req.URL.String()).Should(Equal("https://localhost:1234/test"))

			req, err = botta.HttpRequest("POST", "https://myhost:1234/stuff", "ping")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())

			Expect(req.Method).Should(Equal("POST"))
			Expect(req.URL.String()).Should(Equal("https://myhost:1234/stuff"))
		})
	})
	Context("Get()", func() {
		It("should create a GET http.Request", func() {
			req, err := botta.Get("https://localhost:1234/get")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())
			Expect(req.Method).Should(Equal("GET"))
			Expect(req.URL.String()).Should(Equal("https://localhost:1234/get"))
			expect_body(req, "")
		})
		It("should return an error if unsuccessful", func() {
			req, err := botta.Get("%")
			Expect(err).Should(HaveOccurred())
			Expect(req).Should(BeNil())
		})
	})
	Context("Post()", func() {
		It("should create a POST http.Request", func() {
			req, err := botta.Post("https://localhost:1234/post", "teststring")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())
			Expect(req.Method).Should(Equal("POST"))
			Expect(req.URL.String()).Should(Equal("https://localhost:1234/post"))
			expect_body(req, `"teststring"`)
		})
		It("should return an error if unsuccessful", func() {
			req, err := botta.Post("%", nil)
			Expect(err).Should(HaveOccurred())
			Expect(req).Should(BeNil())
		})
	})
	Context("Put()", func() {
		It("should create a PUT http.Request", func() {
			req, err := botta.Put("https://localhost:1234/put", "testput")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req.Method).Should(Equal("PUT"))
			Expect(req.URL.String()).Should(Equal("https://localhost:1234/put"))
			expect_body(req, `"testput"`)
		})
		It("should return an error if unsuccessful", func() {
			req, err := botta.Put("%", nil)
			Expect(err).Should(HaveOccurred())
			Expect(req).Should(BeNil())
		})
	})
	Context("Patch()", func() {
		It("should create a PATCH http.Request", func() {
			req, err := botta.Patch("https://localhost:1234/patch", "testpatch")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())
			Expect(req.Method).Should(Equal("PATCH"))
			Expect(req.URL.String()).Should(Equal("https://localhost:1234/patch"))
			expect_body(req, `"testpatch"`)
		})
		It("should return an error if unsuccessful", func() {
			req, err := botta.Patch("%", nil)
			Expect(err).Should(HaveOccurred())
			Expect(req).Should(BeNil())
		})
	})
	Context("Delete()", func() {
		It("should create a DELETE http.Request", func() {
			req, err := botta.Delete("https://localhost:1234/delete")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())
			Expect(req.Method).Should(Equal("DELETE"))
			Expect(req.URL.String()).Should(Equal("https://localhost:1234/delete"))
			expect_body(req, "")
		})
		It("should return an error if unsuccessful", func() {
			req, err := botta.Delete("%")
			Expect(err).Should(HaveOccurred())
			Expect(req).Should(BeNil())
		})
	})
	Context("http Request handling", func() {
		var server *httptest.Server

		BeforeEach(func() {
			wd, err := os.Getwd()
			Expect(err).ShouldNot(HaveOccurred())
			server = httptest.NewServer(http.FileServer(http.Dir(wd + "/assets")))
		})
		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		var GET = func(path string, shouldSucceed bool) (*botta.Response, error) {
			req, err := botta.Get(server.URL + path)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(req).ShouldNot(BeNil())

			resp, err := botta.Issue(req)
			if shouldSucceed {
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp).ShouldNot(BeNil())
			} else {
				Expect(err).Should(HaveOccurred())
			}
			return resp, err
		}

		Context("Issue()", func() {
			It("should return errors from http.Client.Do()", func() {
				resp, err := botta.Issue(&http.Request{})
				Expect(err).Should(HaveOccurred())
				Expect(resp).Should(BeNil())
			})
			It("should return a parsed response upon success", func() {
				resp, _ := GET("/test", true)

				Expect(resp.Raw).Should(MatchJSON(`"test successful"`))
				Expect(resp.Data).Should(Equal(interface{}("test successful")))
			})
		})
		Context("ParseResponse()", func() {
			It("should parse json responses successfully", func() {
				resp, _ := GET("/test_json", true)
				Expect(resp.Data).Should(Equal(map[string]interface{}{
					"valid": "json",
				}))
				Expect(resp.HTTPResponse).ShouldNot(BeNil())
				Expect(resp.Raw).Should(MatchJSON(`{"valid":"json"}`))
			})
			It("should parse empty json responses successfully", func() {
				resp, _ := GET("/empty", true)
				Expect(resp).ShouldNot(BeNil())
				Expect(resp.Data).Should(BeNil())
				Expect(resp.HTTPResponse).ShouldNot(BeNil())
				Expect(resp.Raw).Should(Equal([]byte{}))
			})
			It("should return an error for a 200 with invalid json", func() {
				resp, _ := GET("/invalid", false)
				Expect(resp).ShouldNot(BeNil())
				Expect(resp.HTTPResponse).ShouldNot(BeNil())
				Expect(resp.HTTPResponse.StatusCode).Should(Equal(200))
				Expect(resp.Data).Should(BeNil())
				Expect(resp.Raw).Should(Equal([]byte("invalid json\n")))
			})
			It("should return BadResponseCode error with decoded JSON response for status >= 400", func() {
				if server != nil {
					server.Close()
				}
				var jsonHandler http.HandlerFunc
				jsonHandler = func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(404)
					rw.Write([]byte(`{"error":"the server failed to find your content"}`))
				}
				server = httptest.NewServer(jsonHandler)

				resp, err := GET("/not-there-json", false)
				Expect(err.(botta.BadResponseCode).StatusCode).Should(Equal(404))
				Expect(resp.HTTPResponse).ShouldNot(BeNil())
				Expect(resp.Data).Should(Equal(map[string]interface{}{
					"error": "the server failed to find your content",
				}))
				Expect(resp.Raw).Should(MatchJSON(`{"error":"the server failed to find your content"}`))
			})
			It("should return BadResponseCode error with status >= 400", func() {
				resp, err := GET("/not-there-no-json", false)
				Expect(err.(botta.BadResponseCode).StatusCode).Should(Equal(404))
				Expect(resp.HTTPResponse).ShouldNot(BeNil())
				Expect(resp.Data).Should(BeNil())
				Expect(resp.Raw).Should(Equal([]byte("404 page not found\n")))
			})
		})
	})
	Context("Client()", func() {
		It("should return a generic client by default", func() {
			Expect(botta.Client()).Should(Equal(&http.Client{}))
		})
		It("should allow setting a custom http.Client()", func() {
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}
			botta.SetClient(client)

			Expect(botta.Client()).Should(Equal(client))
		})
	})
	Context("BadResponseCode", func() {
		Context("Error()", func() {
			It("returns an error message", func() {
				err := botta.BadResponseCode{
					StatusCode: 123,
					Message:    "this is an error",
					URL:        "https://localhost/test",
				}
				Expect(err.Error()).Should(Equal("https://localhost/test returned 123: this is an error"))

				err = botta.BadResponseCode{
					StatusCode: 321,
					Message:    "this is a different error",
					URL:        "http://asdf.com/",
				}
				Expect(err.Error()).Should(Equal("http://asdf.com/ returned 321: this is a different error"))
			})
		})
	})
})

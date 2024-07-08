package proxy

import (
	"audit-proxy-gateway/internal/config"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type FiberResponseWriter struct {
	*http.Response
	writer io.Writer
}

func (frw *FiberResponseWriter) Header() http.Header {
	return frw.Response.Header
}

func (frw *FiberResponseWriter) WriteHeader(statusCode int) {
	frw.Response.StatusCode = statusCode
}

func (frw *FiberResponseWriter) Write(b []byte) (int, error) {
	return frw.writer.Write(b)
}

func ReverseProxy(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	target := cfg.Application.Proxy.Target

	targetURL, err := url.Parse(target)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Create a new HTTP request for the proxy
	req, err := http.NewRequest(c.Method(), targetURL.String(), io.NopCloser(strings.NewReader(string(c.Body()))))
	if err != nil {
		return err
	}

	// Copy headers from the original request
	req.Header = make(http.Header)
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Set X-Forwarded-For header
	req.Header.Set("X-Forwarded-For", c.IP())

	// Set the original host header
	req.Host = targetURL.Host

	// Custom director to modify the request
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = c.OriginalURL()
		req.URL.RawQuery = string(c.Request().URI().QueryString())
	}

	// Capture the response from the proxy and copy it to the Fiber response
	proxy.ModifyResponse = func(resp *http.Response) error {
		c.Response().SetStatusCode(resp.StatusCode)
		for key, values := range resp.Header {
			for _, value := range values {
				c.Set(key, value)
			}
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		c.Send(body)
		return nil
	}

	// Serve the proxy request using a custom FiberResponseWriter
	frw := &FiberResponseWriter{
		Response: &http.Response{
			Header: http.Header{},
		},
		writer: c.Response().BodyWriter(),
	}

	proxy.ServeHTTP(frw, req)
	return nil
}

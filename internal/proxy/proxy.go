package proxy

import (
	"audit-proxy-gateway/internal/config"
	"audit-proxy-gateway/pkg/logger"
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	cnf := config.GetConfig()
	target := cnf.Application.Proxy.Target

	targetURL, err := url.Parse(target)
	if err != nil {
		logger.GetLogger().Errorf("Failed to parse target URL: %v", err)
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	req, err := createProxyRequest(c, targetURL)
	if err != nil {
		logrus.Errorf("Failed to create proxy request: %v", err)
		return err
	}

	proxy.Director = func(req *http.Request) {
		customDirector(req, c, targetURL)
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		return modifyResponse(c, resp)
	}

	frw := &FiberResponseWriter{
		Response: &http.Response{
			Header: http.Header{},
		},
		writer: c.Response().BodyWriter(),
	}

	proxy.ServeHTTP(frw, req)
	return nil
}

func createProxyRequest(c *fiber.Ctx, targetURL *url.URL) (*http.Request, error) {
	reqBody := bytes.NewReader(c.Body())
	req, err := http.NewRequest(c.Method(), targetURL.String(), reqBody)
	if err != nil {
		return nil, err
	}

	copyHeaders(c, req)
	req.Header.Set("X-Forwarded-For", c.IP())
	req.Host = targetURL.Host

	logger.GetLogger().Info("Request created successfully with body: ", string(c.Body()))
	return req, nil
}

func copyHeaders(c *fiber.Ctx, req *http.Request) {
	req.Header = make(http.Header)
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})
}

func customDirector(req *http.Request, c *fiber.Ctx, targetURL *url.URL) {
	req.URL.Scheme = targetURL.Scheme
	req.URL.Host = targetURL.Host
	req.URL.Path = c.OriginalURL()
	req.URL.RawQuery = string(c.Request().URI().QueryString())
}

func modifyResponse(c *fiber.Ctx, resp *http.Response) error {
	c.Response().SetStatusCode(resp.StatusCode)
	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}
	_, err := io.Copy(c.Response().BodyWriter(), resp.Body)
	if err != nil {
		logger.GetLogger().Errorf("Failed to copy response body: %v", err)
		return err
	}

	logger.GetLogger().Info(
		"Response headers copied successfully. Status code: ", resp.StatusCode, " Headers: ", resp.Header)
	return nil
}

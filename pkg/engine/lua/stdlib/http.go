// ABOUTME: HTTP client module for Lua scripts
// ABOUTME: Provides http.get(), post(), request() functions

package stdlib

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// HTTPConfig holds configuration for the HTTP module
type HTTPConfig struct {
	Timeout         time.Duration
	MaxResponseSize int64
	AllowedSchemes  []string
	UserAgent       string
}

// DefaultHTTPConfig returns a default HTTP configuration
func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Timeout:         30 * time.Second,
		MaxResponseSize: 10 * 1024 * 1024, // 10MB
		AllowedSchemes:  []string{"http", "https"},
		UserAgent:       "llmspell/1.0",
	}
}

// HTTPClient provides HTTP functionality for Lua scripts
type HTTPClient struct {
	config *HTTPConfig
	client *http.Client
}

// NewHTTPClient creates a new HTTP client instance
func NewHTTPClient(config *HTTPConfig) *HTTPClient {
	if config == nil {
		config = DefaultHTTPConfig()
	}

	return &HTTPClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// RegisterHTTP registers the HTTP module with all functions
func RegisterHTTP(L *lua.LState, httpClient *HTTPClient) {
	// Create http module table
	httpModule := L.NewTable()

	// Register functions
	L.SetField(httpModule, "get", L.NewClosure(httpClient.get))
	L.SetField(httpModule, "post", L.NewClosure(httpClient.post))
	L.SetField(httpModule, "request", L.NewClosure(httpClient.request))

	// Register the module
	L.SetGlobal("http", httpModule)
}

// validateURL validates and parses a URL
func (h *HTTPClient) validateURL(urlStr string) (*url.URL, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	schemeAllowed := false
	for _, allowed := range h.config.AllowedSchemes {
		if u.Scheme == allowed {
			schemeAllowed = true
			break
		}
	}
	if !schemeAllowed {
		return nil, fmt.Errorf("scheme %s not allowed", u.Scheme)
	}

	return u, nil
}

// get performs an HTTP GET request
// Usage: content, err = http.get(url)
func (h *HTTPClient) get(L *lua.LState) int {
	urlStr := L.CheckString(1)

	_, err := h.validateURL(urlStr)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	req.Header.Set("User-Agent", h.config.UserAgent)

	resp, err := h.client.Do(req)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer resp.Body.Close()

	// Read response with size limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, h.config.MaxResponseSize))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Check status code
	if resp.StatusCode >= 400 {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)))
		return 2
	}

	L.Push(lua.LString(string(body)))
	return 1
}

// post performs an HTTP POST request
// Usage: response, err = http.post(url, body, content_type)
func (h *HTTPClient) post(L *lua.LState) int {
	urlStr := L.CheckString(1)
	body := L.CheckString(2)
	contentType := L.OptString(3, "application/json")

	_, err := h.validateURL(urlStr)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	req, err := http.NewRequest("POST", urlStr, bytes.NewBufferString(body))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	req.Header.Set("User-Agent", h.config.UserAgent)
	req.Header.Set("Content-Type", contentType)

	resp, err := h.client.Do(req)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer resp.Body.Close()

	// Read response with size limit
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, h.config.MaxResponseSize))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Check status code
	if resp.StatusCode >= 400 {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(respBody))))
		return 2
	}

	L.Push(lua.LString(string(respBody)))
	return 1
}

// request performs a custom HTTP request
// Usage: response, err = http.request({method="GET", url="...", headers={...}, body="..."})
func (h *HTTPClient) request(L *lua.LState) int {
	options := L.CheckTable(1)

	// Extract options
	method := "GET"
	if v := L.GetField(options, "method"); v != lua.LNil {
		method = strings.ToUpper(lua.LVAsString(v))
	}

	urlStr := ""
	if v := L.GetField(options, "url"); v != lua.LNil {
		urlStr = lua.LVAsString(v)
	} else {
		L.Push(lua.LNil)
		L.Push(lua.LString("url is required"))
		return 2
	}

	_, err := h.validateURL(urlStr)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	var body io.Reader
	if v := L.GetField(options, "body"); v != lua.LNil {
		body = bytes.NewBufferString(lua.LVAsString(v))
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Set default User-Agent
	req.Header.Set("User-Agent", h.config.UserAgent)

	// Set custom headers
	if v := L.GetField(options, "headers"); v != lua.LNil {
		if headers, ok := v.(*lua.LTable); ok {
			headers.ForEach(func(key, value lua.LValue) {
				if keyStr, ok := key.(lua.LString); ok {
					req.Header.Set(string(keyStr), lua.LVAsString(value))
				}
			})
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	defer resp.Body.Close()

	// Read response with size limit
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, h.config.MaxResponseSize))
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	// Create response table
	respTable := L.NewTable()
	L.SetField(respTable, "status", lua.LNumber(resp.StatusCode))
	L.SetField(respTable, "body", lua.LString(string(respBody)))

	// Add headers
	headersTable := L.NewTable()
	for key, values := range resp.Header {
		if len(values) > 0 {
			L.SetField(headersTable, key, lua.LString(values[0]))
		}
	}
	L.SetField(respTable, "headers", headersTable)

	L.Push(respTable)
	return 1
}

// RegisterSimpleHTTP registers a simplified HTTP get function (used in examples)
func RegisterSimpleHTTP(L *lua.LState) {
	httpModule := L.NewTable()

	// Simple get function
	L.SetField(httpModule, "get", L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if resp.StatusCode >= 400 {
			L.Push(lua.LNil)
			L.Push(lua.LString(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)))
			return 2
		}

		L.Push(lua.LString(string(body)))
		return 1
	}))

	L.SetGlobal("http", httpModule)
}

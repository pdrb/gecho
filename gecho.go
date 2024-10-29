package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Version
const AppVersion = "1.1.0"

// Usage
const Usage = `Usage: gecho [options]

A simple http "echo" server written in Go

Options:
  -h, --help     Show this help message and exit
  -l, --listen   Listen address (default: ":8090")
  -t, --timeout  Server timeout in seconds (default: 60)
  -v, --version  Show version and exit

Example: gecho --listen 0.0.0.0:80
`

// Model http response
type HTTPResponse struct {
	Data    string            `json:"data"`
	Headers map[string]string `json:"headers"`
	Json    json.RawMessage   `json:"json"`
	Method  string            `json:"method"`
	Origin  string            `json:"origin"`
	Params  map[string]string `json:"params"`
	URL     string            `json:"url"`
}

// Parse headers
func parseHeaders(reqHeaders *http.Header) map[string]string {
	parsedHeaders := make(map[string]string)
	for header, values := range *reqHeaders {
		// We can have multiple headers with the same name, so we iterate over
		// the header string list and create a string with all the headers separated by comma
		for _, value := range values {
			parsedHeaders[header] += value + ","
		}
		// Remove trailling comma from string
		parsedHeaders[header] = strings.TrimSuffix(parsedHeaders[header], ",")
	}
	return parsedHeaders
}

// Parse url query params
func parseParams(query url.Values) map[string]string {
	params := make(map[string]string)
	for param, values := range query {
		// We can have multiple params with the same name, so we iterate over
		// the param string list and create a string with all the param values separated by comma
		for _, value := range values {
			params[param] += value + ","
		}
		// Remove trailling comma from string
		params[param] = strings.TrimSuffix(params[param], ",")
	}
	return params
}

// Mount the complete url, scheme comes from header if it exists
func mountURL(headers map[string]string, req *http.Request) string {
	scheme := "http://"
	proto, ok := headers["X-Forwarded-Proto"]
	if ok {
		scheme = proto + "://"
	}
	url := scheme + req.Host + req.URL.RequestURI()
	return url
}

// Get origin ip from headers or from request remote address
// The order of precedence is (from most to least important):
// Cf-Connecting-Ip > X-Real-IP > X-Forwarded-For (First IP) > Remote Address from Request
func getOrigin(headers map[string]string, remoteAddr *string) string {
	origin := strings.Split(*remoteAddr, ":")[0]

	xForwardedIps, ok := headers["X-Forwarded-For"]
	if ok {
		// Grab only the first ip, this can be a string with multiple ips separated by comma
		origin = strings.Split(xForwardedIps, ",")[0]
	}

	xRealIp, ok := headers["X-Real-Ip"]
	if ok {
		origin = xRealIp
	}

	CfIp, ok := headers["Cf-Connecting-Ip"]
	if ok {
		origin = CfIp
	}

	return origin
}

// Read request body
func readBody(bodyReader io.ReadCloser, w http.ResponseWriter) (string, error) {
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		http.Error(w, "An error occurred while reading the request body", http.StatusInternalServerError)
		slog.Error("error reading the request body", "err", err)
	}
	return string(body), err
}

// Populate json if data is a valid json
func populateJson(data *string) json.RawMessage {
	_, err := json.Marshal(json.RawMessage(*data))
	var jsonData json.RawMessage
	if err == nil {
		jsonData = json.RawMessage(*data)
	}
	return jsonData
}

// Write response to client as a prettified (indented) json
func writeResponse(w http.ResponseWriter, response *HTTPResponse) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	encoder.Encode(*response)
}

// Handle our http request
func mainHandler(w http.ResponseWriter, req *http.Request) {
	// Parse the request headers
	headers := parseHeaders(&req.Header)

	// Parse url params
	params := parseParams(req.URL.Query())

	// Mount the complete url, scheme comes from header if it exists
	url := mountURL(headers, req)

	// Grab origin from headers or from remote address
	origin := getOrigin(headers, &req.RemoteAddr)

	// Read request body
	data, err := readBody(req.Body, w)
	if err != nil {
		return
	}

	// Populate json field
	jsonData := populateJson(&data)

	// Create response
	response := HTTPResponse{
		Data:    data,
		Headers: headers,
		Json:    jsonData,
		Method:  req.Method,
		Origin:  origin,
		Params:  params,
		URL:     url,
	}

	// Write response to client
	writeResponse(w, &response)
}

// Logging middleware
func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		slog.Info("handled request", "method", r.Method, "path", r.URL.Path, "addr", r.RemoteAddr, "elapsed", time.Since(start))
	}
}

func main() {
	// Parse Args
	var listenAddr string
	var timeout int
	var version bool

	flag.StringVar(&listenAddr, "l", ":8090", "listen address")
	flag.StringVar(&listenAddr, "listen", ":8090", "listen address")

	flag.IntVar(&timeout, "t", 60, "timeout")
	flag.IntVar(&timeout, "timeout", 60, "timeout")

	flag.BoolVar(&version, "v", false, "show version")
	flag.BoolVar(&version, "version", false, "show version")

	flag.Usage = func() { fmt.Print(Usage) }
	flag.Parse()

	// Show version and exit
	if version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	// Configure http handler
	http.HandleFunc("/", logRequest(mainHandler))

	// Configure and start server
	fmt.Printf(">>> Starting server at address %v\n", listenAddr)
	srv := &http.Server{
		Addr:         listenAddr,
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

const port = 8090

// Ref - https://stackoverflow.com/a/47286697/11488088
type NotFoundRedirectRespWr struct {
	http.ResponseWriter
	StaticDir string
	status    int
}

func (w *NotFoundRedirectRespWr) WriteHeader(status int) {
	w.status = status
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *NotFoundRedirectRespWr) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	} else {
		// Idea here is to return the contents of the index.html or the default file contents as a response
		// to a file that does not exist, instead of returning a 3xx redirect.
		// This behaviour mimics the functionality of `try_files {}` in Caddy.
		// TODO: Make index.html configurable for the default file
		filePathToReturn := fmt.Sprintf("%sindex.html", w.StaticDir)
		contents, err := os.ReadFile(filePathToReturn)
		if err != nil {
			log.Printf("Found error while returning default page: %v\n", err)
			return len(p), err
		} else {
			// log.Printf("Returning the contents of %s: %s\n", filePathToReturn, string(contents))
			w.ResponseWriter.Header().Add("Content-Type", http.DetectContentType(contents))
			w.ResponseWriter.WriteHeader(http.StatusOK)
			return w.ResponseWriter.Write(contents)
		}
	}
}

// this is to ensure that we always return the index.html so the client side router (like react-router)
// works as usual on the browser
func wrapHandler(h http.Handler, staticDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nfrw := &NotFoundRedirectRespWr{ResponseWriter: w, StaticDir: staticDir}
		h.ServeHTTP(nfrw, r)
	}
}

func replaceEnvInConfig(body []byte) []byte {
	search := regexp.MustCompile(`\$\{([^{}]+)\}`)
	replacedBody := search.ReplaceAllFunc(body, func(b []byte) []byte {
		group1 := search.ReplaceAllString(string(b), `$1`)

		envValue := os.Getenv(group1)
		if len(envValue) > 0 {
			return []byte(envValue)
		} else {
			panic(fmt.Sprintf("Environment variable: %s is not found or value is not set", group1))
		}
	})

	return replacedBody
}

func reverseProxy(route Route) http.HandlerFunc {
	originServerURL, err := url.Parse(route.Url)
	if err != nil {
		log.Fatalf("invalid server URL to proxy: %s", route.Url)
	}

	return func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("[reverse proxy] Processing Request: %s\n", req.RequestURI)
		req.Host = originServerURL.Host
		req.URL.Host = originServerURL.Host
		req.URL.Scheme = originServerURL.Scheme
		req.RequestURI = ""

		// save the response from the origin server
		originServerResponse, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(rw, err)
			return
		}

		// copy the response headers back to the client
		headers := originServerResponse.Header.Clone()
		for k, v := range headers {
			if len(v) == 1 {
				rw.Header().Set(k, v[0])
			} else if len(v) > 1 {
				for _, value := range v {
					rw.Header().Add(k, value)
				}
			}
		}

		// copy the response body to the client
		rw.WriteHeader(originServerResponse.StatusCode)
		io.Copy(rw, originServerResponse.Body)
	}
}

func redirectHandler(redirectURI string) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("[redirect] Redirecting to: %s\n", redirectURI)
		http.Redirect(rw, req, redirectURI, http.StatusFound)
	}
}

func main() {
	configFileName := "./serve.yaml"
	configBytes, err := os.ReadFile(configFileName)
	if err != nil {
		log.Fatalf("Error reading the config file: %s. Error: %v", configFileName, err)
	}
	configFileWithEnvReplaced := replaceEnvInConfig(configBytes)
	fmt.Printf("Resolved Config: \n%s\n", configFileWithEnvReplaced)
	var config Config
	err = yaml.Unmarshal(configFileWithEnvReplaced, &config)
	if err != nil {
		log.Fatalf("Error parsing the config file: %s. \nResolved Config: %s.\n Error: %v", configFileName, configFileWithEnvReplaced, err)
	}

	r := mux.NewRouter()

	// Setup all the reverse proxies
	for _, route := range config.Proxies {
		log.Printf("Adding route: %s to url: %s\n", route.Path, route.Url)
		r.PathPrefix(route.Path).Handler(reverseProxy(route))
	}

	// Setup the redirect handler
	if config.Redirects.RedirectURI != "" {
		log.Printf("Adding redirect handler to: %s\n", config.Redirects.RedirectURI)
		r.PathPrefix("/redirect").Handler(redirectHandler(config.Redirects.RedirectURI))
	} else {
		log.Printf("No redirect_uri specified. Skipping redirect handler setup.\n")
	}

	staticDir := config.Static.Path
	staticRoute := config.Static.Url
	fmt.Printf("Static Dir: %s\n", staticDir)
	fs := http.FileServer(http.Dir(staticDir))
	r.PathPrefix(staticRoute).Handler(http.StripPrefix(staticRoute, wrapHandler(fs, staticDir)))

	log.Printf("Listening on :%d...\n", port)
	server := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf(":%d", port),
		// TODO: See if we should make this configurable
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

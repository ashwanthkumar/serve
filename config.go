package main

type Route struct {
	Path string
	Url  string
}
type Config struct {
	Static  Route `yaml:"static"`
	Proxies []Route `yaml:"proxies"`
	Redirects RedirectConfig `yaml:"redirects"`
}

type RedirectConfig struct {
	RedirectURI string `yaml:"redirect_uri"`
}

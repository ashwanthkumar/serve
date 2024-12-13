package main

type Route struct {
	Path string
	Url  string
}
type Config struct {
	Static  Route `yaml:"static"`
	Proxies []Route `yaml:"proxies"`
}

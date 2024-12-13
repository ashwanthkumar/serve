package main

type Route struct {
	Path string
	Url  string
}
type Config struct {
	Static  Route
	Proxies []Route
}

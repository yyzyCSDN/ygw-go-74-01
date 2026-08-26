package main

import (
	"flag"
	"os"
)

type Config struct {
	Addr    string
	Port    int
	DataDir string
	WebDir  string
}

func LoadConfig() Config {
	return loadConfigFromArgs(os.Args[1:])
}

func loadConfigFromArgs(args []string) Config {
	cfg := Config{Addr: "127.0.0.1", Port: 8090, DataDir: "./data", WebDir: "web"}
	fs := flag.NewFlagSet("cleanroom", flag.ContinueOnError)
	fs.StringVar(&cfg.Addr, "addr", cfg.Addr, "bind address")
	fs.IntVar(&cfg.Port, "port", cfg.Port, "bind port")
	fs.StringVar(&cfg.DataDir, "dir", cfg.DataDir, "data directory")
	fs.StringVar(&cfg.WebDir, "web-dir", cfg.WebDir, "web assets directory")
	_ = fs.Parse(args)
	return cfg
}

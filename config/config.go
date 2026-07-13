package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	TemplatePath string
	UseSyslog    bool
	LocalLogPath string
	SyslogHost   string
	DatabasePath string
	HostAddr     string
}

func DefaultConfig() *Config {
	return &Config{
		TemplatePath: "/var/noodlizer/template",
		UseSyslog:    true,
		SyslogHost:   "localhost:514",
		LocalLogPath: "/var/messages/noodlizer",
		DatabasePath: "/var/noodlizer/db",
		HostAddr:     ":8080",
	}
}

func Load(path string) *Config {
	cdata, err := os.ReadFile(path)
	if err != nil {
		// TODO: log (chicken and egg?)
		panic("unable to load config file: " + path)
	}
	c := &Config{}
	err = json.Unmarshal(cdata, c)
	if err != nil {
		panic("unable to Unmarshal config: " + path)
	}
	return c
}

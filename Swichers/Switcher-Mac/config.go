package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"regexp"
)

type configStructure struct {
	LogLevel string           `json:"log_level"`
	Rules    []*ruleStructure `json:"rules"`
}

type ruleStructure struct {
	Name         string `json:"name"`
	Listen       string `json:"listen"`
	EnableRegexp bool   `json:"enable_regexp"`
	Targets      []*struct {
		Regexp  string         `json:"regexp"`
		regexp  *regexp.Regexp `json:"-"`
		Address string         `json:"address"`
	} `json:"targets"`
	FirstPacketTimeout uint64          `json:"first_packet_timeout"`
	Blacklist          map[string]bool `json:"blacklist"`
}

var config *configStructure

func init() {
	cfgPath := flag.String("config", "./config.json", "config.json file path")
	flag.Parse()

	buf, err := ioutil.ReadFile(*cfgPath)
	if err != nil {
		logrus.Fatalf("failed to load config json: %s", err.Error())
	}

	if err := json.Unmarshal(buf, &config); err != nil {
		logrus.Fatalf("failed to load config json: %s", err.Error())
	}

	if len(config.Rules) == 0 {
		logrus.Fatalf("empty rule")
	}
	lvl, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		logrus.Fatalf("invalid log_level")
	}
	logrus.SetLevel(lvl)

	for i, v := range config.Rules {
		if err := v.verify(); err != nil {
			logrus.Fatalf("verity rule failed at pos %d : %s", i, err.Error())
		}
	}
}

func (c *ruleStructure) verify() error {
	if c.Name == "" {
		return fmt.Errorf("empty name")
	}
	if c.Listen == "" {
		return fmt.Errorf("invalid listen address")
	}
	if len(c.Targets) == 0 {
		return fmt.Errorf("invalid targets")
	}
	if c.EnableRegexp {
		if c.FirstPacketTimeout == 0 {
			c.FirstPacketTimeout = 5000
		}
	}
	for i, v := range c.Targets {
		if v.Address == "" {
			return fmt.Errorf("invalid address at pos %d", i)
		}
		if c.EnableRegexp {
			r, err := regexp.Compile(v.Regexp)
			if err != nil {
				return fmt.Errorf("invalid regexp at pos %d : %s", i, err.Error())
			}
			v.regexp = r
		}
	}
	return nil
}

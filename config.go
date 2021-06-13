package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Config struct {
	LineChannelID          string `yaml:"LineChannelID"`
	LineChannelSecret      string `yaml:"LineChannelSecret"`
	LineChannelAccessToken string `yaml:"LineChannelAccessToken"`
	SlackToken             string `yaml:"SlackToken"`
	SlackChannelId         string `yaml:"SlackChannelId"`
}

func getConfig() *Config {
	c := &Config{}

	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	log.Printf("%s", c)
	return c
}

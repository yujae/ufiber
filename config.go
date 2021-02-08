package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Home                 string
	JwtSigningKey        string
	ActivateKey          string
	AccessKeyExpiredSec  int
	RefreshKeyExpiredSec int
	DB                   []db
	Mail                 []mail
}

type db struct {
	ID         string
	DriverName string
	URL        string
}

type mail struct {
	Host     string
	Port     string
	ID       string
	PW       string
	FromName string
	FromMail string
}

func NewConfig(f string) (Config, error) {
	conf := Config{}
	configFile := f
	filename := configFile
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("'%s' file not found", configFile)
		//return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatalf("'%s' file failed to decode", configFile)
	}
	return conf, err
}

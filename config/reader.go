package config

import (
	"github.com/pelletier/go-toml/v2"
	"io/ioutil"
	"os"
)

func ReadConfig() {
	tomlFile, err := os.Open("config.toml")
	// if we os.Open returns an error then handle it
	if err != nil {
		panic(err)
	}
	// defer the closing of our tomlFile so that we can parse it later on
	defer tomlFile.Close()

	byteValue, _ := ioutil.ReadAll(tomlFile)

	err = toml.Unmarshal([]byte(byteValue), &Config)
	if err != nil {
		panic(err)
	}
}

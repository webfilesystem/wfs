package config

import (
	"io/ioutil"
	"log"
	"strings"

	"os"
	"path/filepath"
)

var path string = filepath.Join(os.Getenv("HOME"), ".wfs")

func GetConfigFiles() ([]os.FileInfo, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func LoadConfigs() (map[string][]byte, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	configs := make(map[string][]byte)
	for _, f := range files {
		file := filepath.Join(path, f.Name())
		if _, err := os.Stat(file); !os.IsNotExist(err) && f.Name() != "lib" {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}
			configs[strings.Replace(f.Name(), ".js", "", -1)] = data
		}
	}
	return configs, nil
}

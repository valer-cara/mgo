package util

import (
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
)

func ParseYaml(filename string) (map[string]interface{}, error) {
	var parsed = make(map[string]interface{})

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &parsed)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

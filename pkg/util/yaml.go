package util

import (
	"errors"
	"fmt"
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
		return nil, errors.New(fmt.Sprintf("YAML Unmarshal error: %s: %v", filename, err))
	}

	return parsed, nil
}

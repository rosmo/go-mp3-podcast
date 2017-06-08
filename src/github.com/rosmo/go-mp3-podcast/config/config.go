package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Configuration struct {
	Channel struct {
		Title       string `yaml:"title"`
		Link        string `yaml:"link"`
		Description string `yaml:"description"`
		Copyright   string `yaml:"copyright"`
		Url         string `yaml:"url"`
		Language    string `yaml:"language"`
	}
	Image struct {
		Title  string `yaml:"title"`
		Url    string `yaml:"url"`
		Itunes string `yaml:"itunes"`
	}
	Items struct {
		Guid struct {
			BaseUrl     string `yaml:"baseUrl"`
			IsPermalink bool   `yaml:"isPermaLink"`
		}
		Link struct {
			BaseUrl string `yaml:"baseUrl"`
		}
		Enclosure struct {
			BaseUrl string `yaml:"baseUrl"`
		}
		Date struct {
			From   string `yaml:"from"`
			Format string `yaml:"format"`
		}
		Filter struct {
			MinimumSize int64 `yaml:"minimumSize"`
		}
	}
	Index struct {
		DateFormat string `yaml:"dateFormat"`
		Template   string `yaml:"template"`
	}
}

func Parse(file string) (*Configuration, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.New("could not read configuration file")
	}

	cfg := Configuration{}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

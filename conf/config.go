package conf

import (
    "gopkg.in/yaml.v3"
    "io/ioutil"
    "log"
)

type Config struct {
    Database struct {
        Path       string `yaml:"path"`
        Migrations struct {
            RunOnStartup bool `yaml:"run_on_startup"`
        } `yaml:"migrations"`
    } `yaml:"database"`
}

func LoadConfig(path string) *Config {
	
    b, err := ioutil.ReadFile(path)
    if err != nil {
        log.Fatalf("failed to read config: %v", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(b, &cfg); err != nil {
        log.Fatalf("failed to unmarshal yaml: %v", err)
    }

    return &cfg
}

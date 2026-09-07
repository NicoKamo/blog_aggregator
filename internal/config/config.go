package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	home_directory, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home_directory, configFileName), nil
}

func write(cfg Config) error {
	JSON, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	json_file_path, err := getConfigFilePath()
	err = os.WriteFile(json_file_path, JSON, 0666)
	if err != nil {
		return err
	}
	return nil
}

func Read() (Config, error) {
	var initConfig Config
	json_file_path, err := getConfigFilePath()
	if err != nil {
		return initConfig, err
	}
	jsonFile, err := os.Open(json_file_path)
	if err != nil {
		fmt.Println(err)
		return initConfig, err
	}
	defer jsonFile.Close()
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return initConfig, err
	}
	err = json.Unmarshal(byteValue, &initConfig)
	if err != nil {
		fmt.Println(err)
		return initConfig, err
	}
	return initConfig, nil
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	err := write(*c)
	if err != nil {
		return err
	}
	return nil
}

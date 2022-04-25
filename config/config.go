// Copyright 2020-present woodsshin. All rights reserved.
// Use of this source code is governed by GNU General Public License v2.0.

package config

import (
	"log"

	"github.com/spf13/viper"
)

// ErrorTolerance ...
type ErrorTolerance struct {
	Count   int    `json:"count"`
	Command string `json:"command"`
}

// BotSettings ...
type BaseSettings struct {
	WebHookUrl            string         `json:"webhookurl"`
	MonitorInterval       int            `json:"interval"`
	RegularReportInterval int            `json:"regularreportinterval"`
	DiscordNotifySnooze   int64          `json:"discordnotifysnooze"`
	WebsocketPort         int64          `json:"websocketport"`
	ErrorTolerance        ErrorTolerance `json:"errortolerance"`
}

// FoundersNodes ...
type NodeSettings struct {
	Name           string   `json:"name"`
	Address        string   `json:"address"`
	Port           uint     `json:"port"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	PrivateKeypath string   `json:"privatekeypath"`
	Nodes          []string `json:"nodes"`
}

// NodeConfig ...
type NodeConfig struct {
	Settings BaseSettings   `json:"settings"`
	Servers  []NodeSettings `json:"servers"`
}

// GetNodeConfig ...
func GetNodeConfig() (NodeConfig, error) {

	var config NodeConfig

	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()
	viper.SetConfigType("json")
	//viper.SetConfigFile("./config/config.json")

	err := viper.ReadInConfig()
	if err != nil {
		return NodeConfig{}, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	if config.Settings.MonitorInterval == 0 {
		config.Settings.MonitorInterval = 90
	}

	if config.Settings.RegularReportInterval == 0 {
		config.Settings.RegularReportInterval = 3600
	}

	if config.Settings.DiscordNotifySnooze == 0 {
		config.Settings.DiscordNotifySnooze = 600
	}

	if config.Settings.WebsocketPort == 0 {
		config.Settings.WebsocketPort = 8080
	}

	if config.Settings.ErrorTolerance.Count == 0 {
		config.Settings.ErrorTolerance.Count = 3
	}

	if config.Servers == nil {
		err = viper.UnmarshalKey("servers", config.Servers)
	}

	for k, v := range config.Servers {
		if len(v.Nodes) == 0 {
			v.Nodes = append(v.Nodes, "founders")
			config.Servers[k] = v
			log.Printf("Monitoring nodes is empty. Added founders as a default to %v(%v)", v.Name, v.Address)
		}
	}

	return config, nil
}

package config

import (
	"log"
	"os"

	"gopkg.in/ini.v1"
)

type ConfigList struct {
	TenantId       string
	ServerEndpoint string
	ServerId       string
	Username       string
	Password       string
	DiscordToken   string
}

var Config ConfigList

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		os.Exit(1)
	}

	Config = ConfigList{
		TenantId:       cfg.Section("conoha").Key("tenantId").String(),
		ServerEndpoint: cfg.Section("conoha").Key("server_endpoint").String(),
		ServerId:       cfg.Section("conoha").Key("serverId").String(),
		Username:       cfg.Section("conoha").Key("username").String(),
		Password:       cfg.Section("conoha").Key("password").String(),
		DiscordToken:   cfg.Section("discord").Key("token").String(),
	}
}

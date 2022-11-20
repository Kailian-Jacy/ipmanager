package cmd

import (
	"flag"
	"fmt"
	config "git.zjuqsc.com/3200100963/ipmanager/Config"
)

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func GetFlags() {
	flag.StringVar(&config.C.ProbePort, "port", "9095", "Port to listen on.")
	flag.StringVar(&config.C.ConfigPath, "config", "", "Configuration file path.")
	flag.Parse()
	if IsFlagPassed("parse") {
		config.C.Mode = "parse"
		fmt.Println("Parsing config file...")
	}
	if IsFlagPassed("help") {
		fmt.Println("Usage: ipm [command] [options].\n\tCommand:\n\t--help Display help and exit.\n\t--parse Parse all the old config before start. Disabled by default.\n\t--port port the port to listen on.")
	}
	if IsFlagPassed("config") {
		fmt.Println("Config path given. Loading config..")
		config.LoadMainConfig(config.C.ConfigPath)
		fmt.Println("Configuration Loaded Successfully.")
	}
}

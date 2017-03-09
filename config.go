/* Copyright © Playground Global, LLC. All rights reserved. */

package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"playground/log"
)

var Debug bool

var configPath string

/* Load unmarshals the contents of a JSON config file into the designated interface, which should
 * generally be a struct. The caller is expected to create a wrapper struct aggregating the config
 * objects of all modules it wants to load; passing such an instance will load the JSON config data
 * into the respective objects. */
func Load(dest interface{}) {
	LoadDirect(configPath, dest)
}

/* LoadDirect is like Load, but loads from the specified config file, bypassing command line
 * parameters. */
func LoadDirect(configFile string, dest interface{}) {
	var err error

	// validate config file input & load its contents if it looks good
	if configFile == "" {
		msg := "-config is required"
		log.Error("config.init", msg)
		panic(msg)
	}
	if configFile, err = filepath.Abs(configFile); err != nil {
		msg := "-config value '"+configFile+"' does not resolve"
		log.Error("config.init", msg)
		panic(msg)
	}
	if stat, err := os.Stat(configFile); (err != nil && !os.IsNotExist(err)) || (stat != nil && stat.IsDir()) {
		msg := "-config value '"+configFile+"' does not stat or is a directory"
		log.Error("config.init", msg, err)
		panic(msg)
	}
	file, err := os.Open(configFile)
	if err != nil {
		msg := "failure opening -config file '"+configFile+"'"
		log.Error("config.init", msg, err)
		panic(msg)
	}
	configContents, err := ioutil.ReadAll(file)
	if err != nil {
		msg := "failure reading -config file '"+configFile+"'"
		log.Error("config.init", msg, err)
		panic(msg)
	}

	// having loaded the raw JSON config data, unmarshal it
	err = json.Unmarshal([]byte(configContents), dest)
	if err != nil {
		// if the error was a JSON syntax error, attempt to report line number it occured at
		if serr, ok := err.(*json.SyntaxError); ok {
			lines := strings.Split(string(configContents), "\n")
			target := int(serr.Offset)
			seen := 0
			for i, line := range lines {
				if target <= (seen + len(line) + 1) { // assume ASCII
					fmt.Println(line)
					msg := "JSON parse error at line "+strconv.Itoa(i+1)+", column "+strconv.Itoa(target-seen)
					log.Error("config.init", msg)
					panic(msg)
				}
				seen += len(line) + 1
			}
		}
		msg := "loading config failed on unmarshal "
		log.Error("config.init", msg, err)
		panic(msg)
	}

	log.Status("config.init", "Config loaded from '"+configFile+"'.")
}

func init() {
	flag.StringVar(&configPath, "config", "", "location of the configuration JSON")
	flag.BoolVar(&Debug, "debug", false, "enable debug logging")

	flag.Parse()
}

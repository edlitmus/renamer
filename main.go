package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"

	yaml "github.com/esilva-everbridge/yaml"
	"github.com/pioz/tvdb"
	yamlv2 "gopkg.in/yaml.v2"
)

func main() {
	// if len(os.Args) <= 2 {
	// 	fmt.Printf("USAGE : %s <target_directory> <target_filename or part of filename> \n", os.Args[0])
	// 	os.Exit(0)
	// }

	config := yaml.New()
	configFile := createConfigPath()
	initConfig(configFile, config)
	c := tvdbClient(config.Get("the_tvdb_api"))

	// tt0056751
	matches, err := c.SearchByImdbID("tt0056751")
	// series, err := c.BestSearch("Doctor Who (1963)")
	if err != nil {
		panic(err)
	}
	series := matches[0]
	err = c.GetSeriesEpisodes(&series, nil)
	if err != nil {
		panic(err)
	}

	for _, e := range series.Episodes {
		if e.AiredSeason == 1 {
			fmt.Printf("S%0.2dE%0.2d %s\n", e.AiredSeason, e.AiredEpisodeNumber, e.EpisodeName)
			name := strings.TrimRightFunc(e.EpisodeName, func(r rune) bool {
				return !unicode.IsLetter(r)
			})
			fmt.Printf("NAME: %s\n", strings.ToLower(name))
		}

		// TODO: need to traverse sXXeXXpXX dirs and rename files as needed after capturing the 'part' number and matching
		// targetDirectory := os.Args[1] // get the target directory
		// fileName := os.Args[2:]       // to handle wildcard such as filename*.go

		// findFile(targetDirectory, fileName)
	}
}

func tvdbClient(tvDBConfig interface{}) tvdb.Client {
	conf := tvDBConfig.(map[interface{}]interface{})
	c := tvdb.Client{Apikey: conf["api_key"].(string), Userkey: conf["user_key"].(string), Username: conf["username"].(string)}
	err := c.Login()
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func createConfigPath() string {
	var usr, _ = user.Current()
	configFile := filepath.Join(usr.HomeDir, ".config/renamer/config.yaml")
	dir := filepath.Dir(configFile)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		log.Printf("error creating config file path: %s", err)
	}

	return configFile
}

func initConfig(configFile string, config *yaml.Yaml) {
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	err = yamlv2.Unmarshal(buf, &config.Values)
	if err != nil {
		log.Fatal(err)
	}
}

// TODO: use this with a walk func to recurse
func findFile(targetDir string, pattern []string) {

	for _, v := range pattern {
		matches, err := filepath.Glob(targetDir + v)

		if err != nil {
			fmt.Println(err)
		}

		if len(matches) != 0 {
			fmt.Println("Found : ", matches)
		}
	}
}

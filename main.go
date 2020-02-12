package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"

	yaml "github.com/esilva-everbridge/yaml"
	"github.com/pioz/tvdb"
	yamlv2 "gopkg.in/yaml.v2"
)

var series tvdb.Series

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("USAGE : %s <target_directory>\n", os.Args[0])
		os.Exit(0)
	}
	targetDirectory := os.Args[1] // get the target directory
	err := os.Chdir(targetDirectory)
	if err != nil {
		panic(err)
	}

	config := yaml.New()
	configFile := createConfigPath()
	initConfig(configFile, config)
	c := tvdbClient(config.Get("the_tvdb_api"))

	// tt0056751 == Doctor Who (1963)
	matches, err := c.SearchByImdbID("tt0056751")
	// series, err := c.BestSearch("Doctor Who (1963)")
	if err != nil {
		panic(err)
	}
	series = matches[0]
	err = c.GetSeriesEpisodes(&series, nil)
	if err != nil {
		panic(err)
	}

	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		re := regexp.MustCompile(`^s(\d+)e(\d+).*?$`)
		if f.IsDir() {
			matches := re.FindStringSubmatch(f.Name())
			s, _ := strconv.Atoi(matches[1])
			ep, _ := strconv.Atoi(matches[2])
			e := findEpisode(s, ep)
			if e == nil {
				continue
			}
			fmt.Printf("s%0.2de%0.2d %s\n", e.AiredSeason, e.AiredEpisodeNumber, e.EpisodeName)
		}
	}
}

func findEpisode(season, episode int) *tvdb.Episode {
	for _, e := range series.Episodes {
		if e.AiredSeason == season && e.AiredEpisodeNumber == episode {
			return &e
		}
	}
	return nil
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

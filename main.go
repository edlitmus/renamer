package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	yaml "github.com/esilva-everbridge/yaml"
	"github.com/pioz/tvdb"
	yamlv2 "gopkg.in/yaml.v2"
)

func main() {
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
		fmt.Printf("S%0.2dE%0.2d %s\n", e.AiredSeason, e.AiredEpisodeNumber, e.EpisodeName)
	}

	// Print the title of the episode 4x08 (season 4, episode 8)
	fmt.Printf(series.GetEpisode(4, 8).EpisodeName)
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

// // findFiles recurses through the given searchDir returning a list of files and it's length
// func findFiles(searchDir string, ext string) ([]string, int) {
// 	fileList := []string{}
// 	searchDir, err := filepath.Abs(searchDir)
// 	if err != nil {
// 		log.Fatal(err)
// 		return fileList, 0
// 	}
// 	err = checkForDir(searchDir)
// 	if err != nil {
// 		log.Fatal(err)
// 		return fileList, 0
// 	}

// 	err = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
// 		if !f.IsDir() {
// 			fileList = append(fileList, path)
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		log.Fatalf("error walking file path: %s", err)
// 	}

// 	return fileList, len(fileList)
// }

// //checkForDir does exactly what it says on the tin
// func checkForDir(filePath string) error {
// 	fi, err := os.Stat(filePath)
// 	if err != nil {
// 		return fmt.Errorf("cannot stat %s: %s", filePath, err)
// 	}
// 	switch mode := fi.Mode(); {
// 	case mode.IsRegular():
// 		return fmt.Errorf("%s is a file", filePath)
// 	case mode.IsDir():
// 		return nil
// 	}

// 	return err
// }

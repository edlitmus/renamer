package main

import (
	"encoding/json"
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

	dirs, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	for _, d := range dirs {
		if d.IsDir() {
			files, _ := findFilesByExt(d.Name(), ".avi")
			for _, f := range files {
				fmt.Printf("FOUND: %s\n", f)
				re := regexp.MustCompile(`s(\d+)e(\d+)p(\d+)\s(.*?)\.avi`)
				matches := re.FindStringSubmatch(f)
				fmt.Printf("MATCHES: %v\n", matches)
				if len(matches) > 0 {
					s, _ := strconv.Atoi(matches[1])
					// ep, _ := strconv.Atoi(matches[2])
					part, _ := strconv.Atoi(matches[3])
					name := fmt.Sprintf("%s (%d)", matches[4], part)
					// e := findEpisode(s, ep)
					e := findEpisodeByName(s, name)
					if e == nil {
						continue
					}
					dump(e)
					// fmt.Printf("s%0.2de%0.2d %s\n", e.AiredSeason, e.AiredEpisodeNumber, e.EpisodeName)
				}
			}
		}
	}
}

// func findEpisode(season, episode int) *tvdb.Episode {
// 	for _, e := range series.Episodes {
// 		if e.AiredSeason == season && e.AiredEpisodeNumber == episode {
// 			return &e
// 		}
// 	}
// 	return nil
// }

func findEpisodeByName(season int, name string) *tvdb.Episode {
	for _, e := range series.Episodes {
		if e.AiredSeason == season && e.EpisodeName == name {
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

//checkForDir does exactly what it says on the tin
func checkForDir(filePath string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %s", filePath, err)
	}
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		return fmt.Errorf("%s is a file", filePath)
	case mode.IsDir():
		return nil
	}

	return err
}

func findFilesByExt(searchDir string, ext string) ([]string, int) {
	fileList := []string{}
	searchDir, err := filepath.Abs(searchDir)
	if err != nil {
		log.Fatal(err)
		return fileList, 0
	}
	err = checkForDir(searchDir)
	if err != nil {
		log.Fatal(err)
		return fileList, 0
	}

	err = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && filepath.Ext(f.Name()) == ext {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal("error walking file path: ", err)
	}

	return fileList, len(fileList)
}

// dump is useful for debugging
func dump(thing interface{}) {
	json, err := json.MarshalIndent(thing, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(json))
}

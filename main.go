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
	"strings"

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

	var files []string
	mkvFiles, count := findFilesByExt("./", ".mkv")
	fmt.Printf("FOUND %d MKV files\n", count)
	files = append(files, mkvFiles...)
	aviFiles, count := findFilesByExt("./", ".avi")
	fmt.Printf("FOUND %d AVI files\n", count)
	files = append(files, aviFiles...)

	// dump(files)
	re := regexp.MustCompile(`s(\d+)e(\d+)p(\d+)\s*(.*?)\.\w{3}$`)
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			panic(err)
		}
		// dump(fileInfo.Name())
		matches := re.FindStringSubmatch(fileInfo.Name())
		if len(matches) > 0 {
			fmt.Printf("MATCHES: %#v\n", matches)
			s, _ := strconv.Atoi(matches[1])
			// part, _ := strconv.Atoi(matches[3])
			// name := fmt.Sprintf("%s (%d)", matches[4], part)
			name := matches[4]
			if strings.Contains(name, "(") {
				name = strings.Split(name, "(")[0]
			}
			if strings.Contains(name, "\\") {
				name = strings.Split(name, "\\")[0]
			}
			name = strings.TrimSpace(name)
			e := findEpisode(s, name)
			if e.Empty() {
				fmt.Println("can't find a match for " + name)
				continue
			}
			// dump(e)
			fileName := fmt.Sprintf("%s - S%0.2dE%0.2d - %s.avi", "Doctor Who (1963)", e.AiredSeason, e.AiredEpisodeNumber, e.EpisodeName)
			err = os.Rename(file, fileName)
			if err != nil {
				panic(err)
			}
			fmt.Printf("RENAMED TO '%s'\n", fileName)
		}
	}

	// dirs, err := ioutil.ReadDir("./")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// for _, d := range dirs {
	// 	if limit == 0 {
	// 		break
	// 	}
	// 	// fmt.Println(d.Name())
	// 	if d.IsDir() && strings.HasPrefix(d.Name(), "s") {
	// 		fmt.Printf("FOUND: %s\n", d.Name())
	// 		re := regexp.MustCompile(`(?i)s(\d+)e(\d+)\s(.*?)$`)
	// 		matches := re.FindStringSubmatch(d.Name())
	// 		fmt.Printf("MATCHES: %#v\n", matches)
	// 		if len(matches) > 0 {
	// 			s, _ := strconv.Atoi(matches[1])
	// 			// ep, _ := strconv.Atoi(matches[2])
	// 			name := matches[3]
	// 			name = strings.Split(name, "(")[0]
	// 			// e := findEpisode(s, ep)
	// 			eps := findEpisodesByName(s, name)
	// 			if len(eps) == 0 {
	// 				fmt.Println("can't find a match for " + name)
	// 				continue
	// 			}
	// 			// dump(eps)
	// 			for _, e := range eps {
	// 				fmt.Printf("DIR MATCH: s%0.2de%0.2d %s\n", e.AiredSeason, e.AiredEpisodeNumber, e.EpisodeName)
	// 			}
	// 		}
	// 		limit--
	// 	}
	// 	//  else {
	// 	// 	fmt.Printf("FOUND FILE: %s\n", d.Name())
	// 	// 	re := regexp.MustCompile(`s(\d+)e(\d+)p(\d+)\s(.*?)\s*\(*.*?$`)
	// 	// 	matches := re.FindStringSubmatch(d.Name())
	// 	// 	fmt.Printf("MATCHES: %v\n", matches)
	// 	// 	if len(matches) > 0 {
	// 	// 		s, _ := strconv.Atoi(matches[1])
	// 	// 		// ep, _ := strconv.Atoi(matches[2])
	// 	// 		part, _ := strconv.Atoi(matches[3])
	// 	// 		name := fmt.Sprintf("%s (%d)", matches[4], part)
	// 	// 		// e := findEpisode(s, ep)
	// 	// 		eps := findEpisodesByName(s, name)
	// 	// 		if len(eps) == 0 {
	// 	// 			fmt.Println("can't find a match for " + name)
	// 	// 			continue
	// 	// 		}
	// 	// 		// dump(eps)
	// 	// 		for _, e := range eps {
	// 	// 			fmt.Printf("FILE MATCH: s%0.2de%0.2d %s\n", e.AiredSeason, e.AiredEpisodeNumber, e.EpisodeName)
	// 	// 		}
	// 	// 	}
	// 	// }
	// }
}

func findEpisode(season int, name string) tvdb.Episode {
	var e tvdb.Episode
	fmt.Printf("LOOKIG FOR '%s'\n", name)
	for _, e := range series.Episodes {
		if e.AiredSeason == season && strings.Contains(e.EpisodeName, name) {
			return e
		}
	}
	return e
}

func findEpisodesByName(season int, name string) []tvdb.Episode {
	eps := make([]tvdb.Episode, 0)
	for _, e := range series.Episodes {
		if e.AiredSeason == season && strings.Contains(e.EpisodeName, name) {
			eps = append(eps, e)
		}
	}
	return eps
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

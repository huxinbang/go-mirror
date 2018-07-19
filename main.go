package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
)

type MirrorConfig struct {
	Version     string `json:"Version"`
	Description string `json:"Description"`
	Mirrors     []struct {
		Source string `json:"Source"`
		Mirror string `json:"Mirror"`
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func main() {
	var config MirrorConfig

	if len(os.Args) <= 2 || os.Args[1] != "get" {
		fmt.Println("usage: go-mirror get [<args>]")
		return
	}
	source := os.Args[len(os.Args)-1]

	go_path := os.Getenv("GOPATH")

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
		return
	}

	home_dir := usr.HomeDir

	if len(go_path) == 0 {
		if len(home_dir) == 0 {
			log.Fatal("GOPATH and HOME environment variable, both not found.")
			return
		} else {
			go_path = home_dir + "/go"
			log.Printf("setting GOPATH=%s\n", go_path)
		}
	}

	config_path := home_dir + "/.go-mirror/config.json"
	// config_path := "config.json"

	if !pathExists(config_path) {
		log.Printf("can't find %s\n", config_path)
		return
	} else {
		raw, err := ioutil.ReadFile(config_path)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		json.Unmarshal(raw, &config)
	}

	for _, cfg := range config.Mirrors {
		if strings.HasPrefix(source, cfg.Source) {
			mirror_path := go_path + "/src/" + cfg.Mirror
			source_path := go_path + "/src/" + cfg.Source
			source_parent_path := path.Dir(source_path)

			if _, err := os.Stat(mirror_path); os.IsNotExist(err) {
				os.Args[len(os.Args)-1] = cfg.Mirror
				mirror_args := os.Args[1:len(os.Args)]
				log.Print(">Exec test: go ", mirror_args)
				cmd := exec.Command("go", mirror_args...)
				cmd.Run()
				// stdout, _ := cmd.CombinedOutput()
				// log.Printf(string(stdout))

			} else {
				log.Printf(">Info: mirror already exists: %s\n", mirror_path)
			}

			if _, err := os.Stat(source_parent_path); os.IsNotExist(err) {
				log.Printf(">Exec: mkdir -p %s\n", source_parent_path)
				cmd := exec.Command("mkdir", "-p", source_parent_path)
				stdout, err := cmd.CombinedOutput()

				if err != nil {
					println(err.Error())
					return
				}
				print(string(stdout))
			} else {
				log.Printf(">Info: source parent already exists: %s\n", source_parent_path)
			}

			if _, err := os.Stat(source_path); os.IsNotExist(err) {
				log.Printf(">Exec: ln -s %s %s\n", mirror_path, source_path)
				cmd := exec.Command("ln", "-s", mirror_path, source_path)
				stdout, err := cmd.CombinedOutput()

				if err != nil {
					println(err.Error())
					return
				}
				print(string(stdout))
			} else {
				log.Printf(">Info: source already exists: %s\n", source_path)
			}
			return
		}
	}

	fmt.Printf("!!!no mirror config for %s\n", source)
}

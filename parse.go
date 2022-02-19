package main

import (
	"fmt"
	"os"
	"log"
	"bufio"
	"path/filepath"
	"github.com/BurntSushi/toml"
)


type tomlConfig struct {
	Name string 
	Feeds []tomlFeed `toml:"feeds"`
}


type tomlFeed struct {
	Name string
	Link string
}

func parseTextFiles( matches []string, tasks []feedTask ) []feedTask {
	
	for _,urlfile := range matches {

		fmt.Printf( "Parsing text file %s.\n", urlfile )
		
		_, filename := filepath.Split( urlfile )
		catname := remove_ext( filename )

		file, err := os.Open( urlfile )
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		var c int
		for scanner.Scan() {

			tasks = append( tasks, feedTask{Link: scanner.Text(), Category: catname} )
			c = c+1
		} //all urls in file

		fmt.Printf( "Found %d feeds.\n", c)
		
	} //all files

	return tasks
}

func parseTomlFiles( matches []string, tasks []feedTask ) []feedTask {
	
	for _,tomlfile := range matches {

		fmt.Printf( "Parsing toml file %s.\n", tomlfile )

		_, filename := filepath.Split( tomlfile )
		catname := remove_ext( filename )

		var config tomlConfig
		_,err := toml.DecodeFile( tomlfile, &config )
		if err != nil {
			log.Print(err)
			continue
		}
		
		//replace category name
		if len(config.Name) > 0 {
			catname = config.Name
		}

		fmt.Printf( "Found %d feeds.\n", len(config.Feeds) )

		for _,tfeed := range config.Feeds {
			tasks = append( tasks,
				feedTask{Name: tfeed.Name, Link: tfeed.Link, Category: catname} )
		}
	}

	return tasks
}

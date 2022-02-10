package main

import (
	"fmt"
	"sort"
	"html/template"
	"net/http"
	"github.com/pkg/browser"
	"github.com/mmcdole/gofeed"
	"path/filepath"
	"log"
	"bufio"
	"os"
	"strings"
	"path"
	"time"
)

func remove_ext(fn string) string {
      return strings.TrimSuffix(fn, path.Ext(fn))
}

func time_since(item *gofeed.Item) time.Duration {
	return time.Now().Sub( *item.PublishedParsed )
}

func main() {

	var data map[string][]*gofeed.Feed
	data = make( map[string][]*gofeed.Feed )

	matches, _ := filepath.Glob( "config/*.txt" )
	fmt.Println( matches )

	fp := gofeed.NewParser()

	for _,urlfile := range matches {

		_, filename := filepath.Split( urlfile )
		catname := remove_ext( filename )

		file, err := os.Open( urlfile )
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		
		for scanner.Scan() {
			//fmt.Println(scanner.Text())

			//try to parse each line as a URL
			feed, err := fp.ParseURL( scanner.Text() )
			
			if err != nil {
				log.Println(err)
				continue
			}

			//if successfully parsed, add to data

			//this does ascending sort on the feed
			sort.Sort( sort.Reverse( feed ))

			//pass max 10 items
			feed.Items = feed.Items[0:10]

			//annotate items with time_since
			for _, item := range feed.Items {
				m := make(map[string]string)
				m["time_since"] = fmt.Sprint( time_since( item ) )
				item.Custom = m
			}

			//add to slice
			data[catname] = append( data[catname], feed )

		} //all urls in file
	} //all files

	//fmt.Println( data )

	funcMap := template.FuncMap{
        "now": time.Now,
		"time_since": time_since,
    }
	
    tmpl := template.Must(template.New("tabs.html").Funcs(funcMap).ParseFiles("tabs.html"))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    tmpl.Execute(w, data)
    })

	go browser.OpenURL("http://localhost:8080")
	
	http.ListenAndServe(":8080", nil)
}



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
	"embed"
	"github.com/rickb777/date/period"
	"github.com/BurntSushi/toml"
	"context"
)

func remove_ext(fn string) string {
      return strings.TrimSuffix(fn, path.Ext(fn))
}

func format_time_since( t1 time.Time, t2 time.Time ) (string,string) {

	defer func() {
        if r := recover(); r != nil {
            //fmt.Println("Recovered in f", r)
			//return "eternity", "blue"
        }
    }()
	
	p := period.Between( t1, t2 ).Normalise(false)

	//todo: find a way to recover from panic

	var str, cstr string
	if p.Years() > 0 {
		str = fmt.Sprintf( "%dY", p.Years() )
		cstr = "BurlyWood"
	} else if p.Months() > 0 {
		str = fmt.Sprintf( "%dM", p.Months() )
		cstr = "BurlyWood"
	} else if p.Days() > 0 {
		str = fmt.Sprintf( "%dD", p.Days() )
		cstr = "CadetBlue"
	} else if p.Hours() > 0 {
		str = fmt.Sprintf( "%dh", p.Hours() )
		cstr = "green"
	} else if p.Minutes() > 0 {
		str = fmt.Sprintf( "%dm", p.Minutes() )
		cstr = "green"
	} else {
		str = fmt.Sprintf( "%ds", p.Seconds() )
		cstr = "green"
	}
	return str, cstr		
}

//go:embed tabs.html
var tabs embed.FS

//parse and process a feed from a URL with optional name to replace the title
func process_feed( fp *gofeed.Parser, url string, name string ) (*gofeed.Feed, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	feed, err := fp.ParseURLWithContext( url, ctx )

	//if we have a failure to parse, return the error
	if err != nil {
		return nil, err
	}

	ddate := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

	//make sure each feed has PublishedParsed
	for _,item := range feed.Items {
		if item.PublishedParsed == nil {
			//give items a default date
			item.PublishedParsed = &ddate
		}
	}

	//sort the feed in ascending order
	sort.Sort( sort.Reverse( feed ))
	
	//pass max 10 items

	num_items := len( feed.Items )
	if num_items >= 10 {
		feed.Items = feed.Items[0:10]
	}
		
	//annotate items with time_since
	for _, item := range feed.Items {
		m := make(map[string]string)

		str,cstr := format_time_since( *item.PublishedParsed, time.Now() )
		m["time_since"] = str
		m["time_color"] = cstr
		item.Custom = m
	}

	//replace feed title if nonempty
	if len(name) > 0 {
		feed.Title = name
	}

	//return the feed
	return feed, nil
}

func sort_category( catlist []*gofeed.Feed ) {

	//sort items in category
	sort.Slice( catlist, func(i, j int) bool {
		var t1, t2 time.Time
		t1 = *catlist[i].Items[0].PublishedParsed
		t2 = *catlist[j].Items[0].PublishedParsed
		return t2.Before( t1 )
	})
}

type tomlConfig struct {
	Name string 
	Feeds []tomlFeed `toml:"feeds"`
}


type tomlFeed struct {
	Name string
	Link string
}


func main() {

	fmt.Println( "Welcome to owlmoon." )

	var data map[string][]*gofeed.Feed
	data = make( map[string][]*gofeed.Feed )

	fp := gofeed.NewParser()

	//check for existence of "feeds" in working directory
	fmt.Print( "Checking existence of \"feeds\" in working directory... " )

	_, err := os.Stat( "feeds" )
	if err != nil {
		fmt.Println( "not found." )
	} else {
		fmt.Println( "found. ")
	}

	//PARSE text files
	matches, _ := filepath.Glob( "feeds/*.txt" )

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
		
		for scanner.Scan() {

			feed, err := process_feed( fp, scanner.Text(), "" )

			if err != nil {
				log.Println(err)
				continue
			}
			
			//add to slice
			data[catname] = append( data[catname], feed )

		} //all urls in file

		fmt.Printf( "Found %d feeds.\n", len(data[catname]))

		sort_category( data[catname] )
		
	} //all files

	//PARSE toml files
	matches, _ = filepath.Glob( "feeds/*.toml" )

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
		
		//fmt.Println( config )
		//fmt.Println( mdata )
		
		//replace category name
		if len(config.Name) > 0 {
			catname = config.Name
		}

		fmt.Printf( "Found %d feeds.\n", len(config.Feeds) )

		for _,tfeed := range config.Feeds {

			feed, err := process_feed( fp, tfeed.Link, tfeed.Name )

			if err != nil {
				log.Println(err)
				continue
			}
			
			//add to slice
			data[catname] = append( data[catname], feed )
		}

		sort_category( data[catname] )
	}
	
    tmpl := template.Must(template.ParseFS(tabs, "*.html"))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    tmpl.Execute(w, data)
    })

	go browser.OpenURL("http://localhost:8080")

	fmt.Println( "Ctrl-C or close console to stop http server." )
	
	http.ListenAndServe(":8080", nil)
}



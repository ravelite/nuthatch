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
	"github.com/rickb777/date/period"
)

func remove_ext(fn string) string {
      return strings.TrimSuffix(fn, path.Ext(fn))
}

// func time_since(item *gofeed.Item) time.Duration {
// 	return time.Now().Sub( *item.PublishedParsed )
// }

func format_time_since( t1 time.Time, t2 time.Time ) (string,string) {
	p := period.Between( t1, t2 ).Normalise(false)

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
				//m["time_since"] = fmt.Sprintf("duration: %s", time_since(item).Round(time.Second) )

				str,cstr := format_time_since( *item.PublishedParsed, time.Now() )
				m["time_since"] = str
				m["time_color"] = cstr
				item.Custom = m
			}

			//add to slice
			data[catname] = append( data[catname], feed )

		} //all urls in file

		catlist := data[catname]

		//sort items in category
		sort.Slice( catlist, func(i, j int) bool {
			var t1, t2 time.Time
			t1 = *catlist[i].Items[0].PublishedParsed
			t2 = *catlist[j].Items[0].PublishedParsed
			return t2.Before( t1 )
		})

		data[catname] = catlist	
		
	} //all files

	//fmt.Println( data )

	funcMap := template.FuncMap{
        "now": time.Now,
		//"time_since": time_since,
    }
	
    tmpl := template.Must(template.New("tabs.html").Funcs(funcMap).ParseFiles("tabs.html"))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    tmpl.Execute(w, data)
    })

	go browser.OpenURL("http://localhost:8080")
	
	http.ListenAndServe(":8080", nil)
}



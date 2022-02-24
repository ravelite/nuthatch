package main

import (
	"fmt"
	"sort"
	"html/template"
	"net/http"
	"path/filepath"
	"log"
	"os"
	//"strings"
	//"path"
	"time"
	"embed"
	"context"
	"github.com/pkg/browser"
	"github.com/mmcdole/gofeed"
	//"github.com/rickb777/date/period"
	"github.com/shibukawa/configdir"
)

//parse and process a feed from a URL with optional name to replace the title
//this is the main work to be done for each feed
func process_feed( fp *gofeed.Parser, url string, name string ) (*gofeed.Feed, error) {

	//this version is for using non-default timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	feed, err := fp.ParseURLWithContext( url, ctx )

	//without timeout
	//feed, err := fp.ParseURL( url )
	
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

//embed the tabs template as a file descriptor
//go:embed tabs.html
var tabs embed.FS

//task structure used to collect feed URLs and turn them into parsed feeds
type feedTask struct {
	Name string
	Link string
	Category string
	Feed *gofeed.Feed
}

func worker( id int, tasks []feedTask, jobs <-chan int, results chan<- int ) {
	fp := gofeed.NewParser()

	for j := range jobs {
		feed, err := process_feed( fp, tasks[j].Link, tasks[j].Name )

		if err != nil {
			log.Println(err)
		} else {
			tasks[j].Feed = feed
		}
		//consider the job finished
		results <- j
	}
}

func main() {

	fmt.Println( "Welcome to owlmoon." )

	var data map[string][]*gofeed.Feed
	data = make( map[string][]*gofeed.Feed )

	//fp := gofeed.NewParser()

	//check for existence of "feeds" in working directory
	fmt.Print( "Checking existence of \"feeds\" in working directory... " )

	_, err := os.Stat( "feeds" )
	if err != nil {
		fmt.Println( "not found." )
	} else {
		fmt.Println( "found. ")
	}

	//check for existence of "feeds" in user config directly
	configDirs := configdir.New("ravelite", "nuthatch")
	configDirs.LocalPath, _ = filepath.Abs(".") //add local path

	// Checks to user folder
	folders := configDirs.QueryFolders(configdir.Global)
	//folders[0].WriteFile("setting.json", data)
	userFolder := folders[0]

	fmt.Print( "Checking existence of \"feeds\" in user configuration directory:\n" )
	fmt.Print( userFolder.Path )

	_, err = os.Stat( filepath.Join(userFolder.Path, "feeds" ))
	if err != nil {
		fmt.Println( "... not found." )
	} else {
		fmt.Println( "... found. " )
	}
	
	//folder := configDirs.QueryFolderContainsFile("setting.json")

	var tasks []feedTask

	//PARSE text files
	matches, _ := filepath.Glob( "feeds/*.txt" )
	tasks = parseTextFiles( matches, tasks )

	//PARSE toml files
	matches, _ = filepath.Glob( "feeds/*.toml" )
	tasks = parseTomlFiles( matches, tasks )


	fmt.Printf( "Total feeds to fetch: %d\n", len( tasks ) )

	//use workers to fetch feeds
	numJobs := len( tasks )
	jobs := make(chan int, numJobs)
    results := make(chan int, numJobs)

	//setup workers
	numWorkers := 10
	for w := 0; w < numWorkers; w++ {
        go worker(w, tasks, jobs, results)
    }

	//setup jobs
	for j := 0; j < numJobs; j++ {
        jobs <- j
    }
    close(jobs)

	//wait on results
	for a := 0; a < numJobs; a++ {
        <-results
    }

	//add all feeds to categories
	for _,ftask := range tasks {
		if ftask.Feed != nil {
			data[ftask.Category] = append( data[ftask.Category], ftask.Feed )
		}
	}
	
	//finally, sort all categories
	for cat := range data {
		sort_category( data[cat] )
	}
	
    tmpl := template.Must(template.ParseFS(tabs, "*.html"))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	    tmpl.Execute(w, data)
    })

	go browser.OpenURL("http://localhost:8080")

	fmt.Println( "Ctrl-C or close console to stop http server." )
	
	http.ListenAndServe(":8080", nil)
}



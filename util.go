package main

import (
	"strings"
	"time"
	"path"
	"fmt"
	"sort"
	"github.com/rickb777/date/period"
	"github.com/mmcdole/gofeed"
)

//remove the extension from a filename
func remove_ext(fn string) string {
      return strings.TrimSuffix(fn, path.Ext(fn))
}

//format the time between a post and now as:
//1. the most significant time unit, abbreviated
//2. an HTML color (for now), could be replaced with css classes
//these are used in the template rendering
func format_time_since( t1 time.Time, t2 time.Time ) (string,string) {

	//recover if period.Between panics on improper Period
	defer func() {
        if r := recover(); r != nil {
            //fmt.Println("Recovered in f", r)
			//return "eternity", "blue"
        }
    }()
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

//how to sort a list of feeds, when all have a published date
func sort_category( catlist []*gofeed.Feed ) {

	//sort items in category
	sort.Slice( catlist, func(i, j int) bool {
		var t1, t2 time.Time
		t1 = *catlist[i].Items[0].PublishedParsed
		t2 = *catlist[j].Items[0].PublishedParsed
		return t2.Before( t1 )
	})
}


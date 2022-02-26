package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"path/filepath"
	"github.com/gilliek/go-opml/opml"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {

	//usage goes here
	if len( os.Args ) < 2 {
		fmt.Println( "usage: nutconvert [feeds.opml]" )
		return
	}

	arg1 := os.Args[1]
	
	doc, err := opml.NewOPMLFromFile( arg1 )
	if err != nil {
		log.Fatal(err)
	}

	//extract list of categories with slices of outlines
	var catset map[string][]opml.Outline
	catset = make( map[string][]opml.Outline )
	
	for _,outline := range doc.Body.Outlines {
		cats := strings.Split( outline.Category, "," )
		for _,cat := range cats {
			catset[cat] = append( catset[cat], outline )
		}
	}

	os.MkdirAll("opml_convert", os.ModePerm)

	for k,v := range catset {

		fname := fmt.Sprintf( "%s.toml", k )

		//from fname, replace difficult characters
		fname_safe := strings.Replace( fname, "/", "-", -1 )
		fmt.Println( fname_safe )
		
		fpath := filepath.Join( "opml_convert", fname_safe )
		f, err := os.Create( fpath )
		check(err)
		defer f.Close()

		//give the full category name
		catline := fmt.Sprintf( "name = \"%s\"\n\n", k )
		f.WriteString( catline )

		//outlines in category
		for _,outline := range v {

			f.WriteString( "[[feeds]]\n" );
			
			nameline := fmt.Sprintf( "name = \"%s\"\n", outline.Title )
			f.WriteString( nameline )

			linkline := fmt.Sprintf( "link = \"%s\"\n", outline.XMLURL )
			f.WriteString( linkline )

			f.WriteString( "\n" )
		}
	}
}

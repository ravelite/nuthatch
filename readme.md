
# Nuthatch Feedspeeder

This is a minimal/hobby feedreader, inspired by [fraidycat](https://github.com/kickscondor/fraidycat).
But a bit more manual.

First, user subscriptions are parsed from user configuration files.
The feeds are fetched just once as a batch, and compiled into a static browseable report,
which is served to `localhost:8080`.

## Building

You will need a standard Go installation. 

From the repo directory, `go mod tidy` should fetch the dependencies,
and `go build .` should build the main executable,
while `go install .` should build and install the main executable alongside your Go installation.

If the OPML converter is needed, descend to the `nutconvert` directory and build it.

## Setup: user subscriptions

In order to use `nuthatch` as a feed reader, we first need to setup our subscriptions.

Here we break with tradition, and ditch OMPL files for simple config files.
In doing this, we elide the need for a GUI to manage user subscriptions.
This should be manageable, at least for a certain type of user.

Nonetheless, I try to provide a simple converter `nutconvert` for your current OPML files, described below.

User subscriptions are organized as flat containers called categories, with a config file for each. 

Nuthatch will search for config files in a `feeds` subdirectory in two locations,
either the current directory or in a platform-dependent user configuration directory.
The latter is reported by the program when it is run.

Files can be either simple text files with a `txt` extention, with a feed URL on each line:

```
https://www.site1.com/feed.xml
https://www.site2.com/atom.rss
```

in which case, the category be named after the base filename,

or toml files with a `toml` extension:

```toml
name = "My Category"

[[feeds]]
name = "Mr. Muster's Mysteries"
link = "https://muster.com/feed.xml"

[[feeds]]
name = "Site 2"
link = "https://www.site2.com/atom.rss"

# more feeds below

```
where an optional `name` variable can rename the category.
Each feed block should have a `link` property and can optionally be renamed with `name`.

## Converting an OPML file to toml

To attempt to convert an OPML file, run `nutconvert [file.opml]`.
If successful, this will output converted feeds in an `opml_convert` subdirectory.

This will only regard the `title`, `link`, and `category` channel properties.
It assumes that `category` may be a comma separated list of tags, 
and tries to collect these in separate category files.

If this works, you can copy these to a `feeds` directory to use,
and possibly edit later to refine them.

## Using

To use, simply run the main `nuthatch` executable. 
It should parse your subscriptions, fetch the feeds, and open up a browser to `localhost:8080`.

Categories are displayed as tabs and each feed is a `details` element that can be expanded.

Once the page is loaded, the program can be closed immediately.




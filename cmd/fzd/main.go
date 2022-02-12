package main

import (
	"os"

	"github.com/horacehylee/fzd"
)

func main() {
	log := newLogger()
	if len(os.Args) < 2 {
		log.Fatal("search term should be provided as 2nd arg")
	}
	term := os.Args[1]
	log.Infof("search for \"%v\"", term)

	c, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("config: %+v", c)

	var options []fzd.IndexerOption
	for _, l := range c.Locations {
		path := os.ExpandEnv(l.Path)
		locationOption := fzd.LocationOption{
			Filters: l.Filters,
			Ignores: l.Ignores,
		}
		options = append(options, fzd.WithLocation(path, locationOption))
	}

	indexer, err := fzd.NewIndexer("fzd.bleve", options...)
	if err != nil {
		log.Fatal(err)
	}
	defer indexer.Close()

	err = indexer.Index()
	if err != nil {
		log.Fatal(err)
	}

	count, err := indexer.DocCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("files count: %v", count)

	res, err := indexer.Search(term)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("result: %+v", res)
}

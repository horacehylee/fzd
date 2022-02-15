package main

import (
	"errors"
	"os"

	"github.com/horacehylee/fzd"
	"github.com/sirupsen/logrus"
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
		locationOption := fzd.LocationOption{
			Filters: l.Filters,
			Ignores: l.Ignores,
		}
		options = append(options, fzd.WithLocation(l.Path, locationOption))
	}

	indexer, err := fzd.NewIndexer(c.Index.BasePath, options...)
	if err != nil {
		log.Fatal(err)
	}

	err = indexer.Open()
	if err != nil {
		if errors.Is(err, fzd.ErrIndexHeadDoesNotExist) {
			log.Info("==============")
			log.Info("Setup Index")
			log.Info("==============")
			name, err := indexer.Index()
			if err != nil {
				log.Fatal(err)
			}
			err = indexer.OpenAndSwap(name)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}

	log.Info("==============")
	log.Info("Before Reindex")
	log.Info("==============")
	check(indexer, log, term)

	name, err := indexer.Index()
	if err != nil {
		log.Fatal(err)
	}
	err = indexer.OpenAndSwap(name)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("==============")
	log.Info("After Reindex")
	log.Info("==============")
	check(indexer, log, term)
}

func check(indexer *fzd.Indexer, log *logrus.Logger, term string) {
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

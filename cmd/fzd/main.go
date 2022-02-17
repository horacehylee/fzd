package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/horacehylee/fzd"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

func main() {
	log := newLogger()
	c, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}

	app := &cli.App{
		Usage: "Golang file indexer and fuzzy file finder utiliy tool",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "num",
				Aliases: []string{"n"},
				Value:   5,
				Usage:   "Number of results",
			},
		},
		Action: func(ctx *cli.Context) error {
			indexer, err := newIndexer(c)
			if err != nil {
				return err
			}
			switch ctx.NArg() {
			case 0:
				return statusOrIndex(ctx, c, indexer)
			case 1:
				return search(ctx, c, indexer)
			default:
				return fmt.Errorf("too much arguments are passed: %v", ctx.Args())
			}
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func statusOrIndex(ctx *cli.Context, c config, indexer *fzd.Indexer) error {
	err := indexer.Open()
	if err != nil {
		return indexIfNotExists(indexer, err)
	}
	t, err := indexer.LastIndexed()
	if err != nil {
		return err
	}
	fmt.Printf("Index was last indexed at %v\n", t.Format("2006-01-02 15:04"))
	yes := yesNo("Do you want to reindex it now")
	if !yes {
		return nil
	}
	return index(indexer)
}

func indexIfNotExists(indexer *fzd.Indexer, err error) error {
	if err == nil {
		return nil
	}
	if !errors.Is(err, fzd.ErrIndexHeadDoesNotExist) {
		return err
	}
	fmt.Println("Index is not created yet")
	yes := yesNo("Do you want to create it now")
	if !yes {
		return nil
	}
	return index(indexer)
}

func index(indexer *fzd.Indexer) error {
	name, err := indexer.Index()
	if err != nil {
		return err
	}
	err = indexer.OpenAndSwap(name)
	if err != nil {
		return err
	}
	count, err := indexer.DocCount()
	if err != nil {
		return err
	}
	fmt.Printf("Indexed for %v files\n", count)
	return nil
}

func search(ctx *cli.Context, c config, indexer *fzd.Indexer) error {
	term := ctx.Args().First()
	if term == "" {
		return fmt.Errorf("term cannot be blank")
	}
	err := indexer.Open()
	if err != nil {
		err = indexIfNotExists(indexer, err)
		if err != nil {
			return err
		}
	}
	res, err := indexer.Search(term)
	if err != nil {
		return err
	}
	num := ctx.Int("num")
	for i, h := range res.Hits {
		if i >= num {
			break
		}
		fmt.Printf("%v\n", h.ID)
	}
	return nil
}

func yesNo(msg string) bool {
	prompt := promptui.Prompt{
		Label:     msg,
		IsConfirm: true,
	}
	_, err := prompt.Run()
	return err == nil
}

func newIndexer(c config) (*fzd.Indexer, error) {
	var options []fzd.IndexerOption
	for _, l := range c.Locations {
		locationOption := fzd.LocationOption{
			Filters: l.Filters,
			Ignores: l.Ignores,
		}
		options = append(options, fzd.WithLocation(l.Path, locationOption))
	}
	return fzd.NewIndexer(c.Index.BasePath, options...)
}

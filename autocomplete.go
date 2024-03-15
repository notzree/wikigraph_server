package main

import (
	d "github.com/notzree/wikigraph_server/database"
)

type AutoCompleter interface {
	AutoComplete(prefix string) ([]string, error)
}

type WikigraphAutoCompleter struct {
	db d.LookupHandler
}

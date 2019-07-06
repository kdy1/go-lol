package main

import (
	"go/format"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/go-lol/lol/go-lol-generator/lolgen"
	"github.com/go-lol/lol/go-lol-generator/lolregi"
)

func init() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetLevel(log.InfoLevel)
}

const targetFile = "api.gen.go"

func main() {
	reg := lolregi.NewDefault()
	reg.PrintDebugInfo()

	g := lolgen.New(reg)
	src, err := formatFile(targetFile, g.Generate())
	if err != nil {
		log.Fatalf("Failed to format generated file. %v", err)
		return
	}

	if err := ioutil.WriteFile(targetFile, src, 0644); err != nil {
		log.Fatalf("Failed to write to %s\nError: %v", targetFile, err)
		return
	}
}

func formatFile(filename string, src []byte) ([]byte, error) {
	data, err := format.Source(src)
	if err != nil {
		return nil, err
	}

	return data, nil
}

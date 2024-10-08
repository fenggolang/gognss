// sitelogs2sta reads IGS sitelog files and uses them to generate a Bernese
// Station Information (STA) file.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/de-bkg/gognss/pkg/site"
)

const (
	version = "0.0.8"
)

func main() {
	fmtVers := ""
	fs := flag.NewFlagSet("sitelogs2sta/"+version, flag.ExitOnError)
	fs.StringVar(&fmtVers, "fmtvers", "1.03", "The STA-File format version. Supported versions: 1.01, 1.03.")
	fs.Usage = func() {
		fmt.Println("sitelogs2sta - create a Bernese STA-File based on IGS sitelog formated files")
		fmt.Println("")
		fmt.Println("USAGE: sitelogs2sta [OPTIONS] FILE...")
		fmt.Printf("\nFLAGS:\n")
		fs.PrintDefaults()
		fmt.Println(`
EXAMPLES:
    $ sitelogs2sta ~/sitelogs/*.log >out.STA 2>out.sta.err
        `)

		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Source: %s\n", "https://github.com/de-bkg/gognss/tree/master/cmd/sitelogs2sta")
		fmt.Println("BKG Frankfurt, 2022")
	}
	fs.Parse(os.Args[1:])

	var sites site.Sites
	for _, slPath := range fs.Args() {
		s, err := readSL(slPath)
		if err != nil {
			log.Printf("E! %s: %v", slPath, err)
			continue
		}
		sites = append(sites, s)
	}

	err := sites.WriteBerneseSTA(os.Stdout, fmtVers)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func readSL(slPath string) (*site.Site, error) {
	f, err := os.Open(slPath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer f.Close()

	s, err := site.DecodeSitelog(f)
	if err != nil {
		return nil, fmt.Errorf("decode sitelog: %v", err)
	}
	for _, warn := range s.Warnings {
		log.Printf("WARN: %s: %v", slPath, warn)
	}

	err = s.ValidateAndClean(false)
	if err != nil {
		return nil, fmt.Errorf("validate sitelog: %v", err)
	}

	// Try to get the NineCharID from the filename.
	// This is necessary until the the NineCharID is official part of the sitelog.
	if s.Ident.NineCharacterID == "" {
		if nineCharID := site.IDByFilename(filepath.Base(slPath)); len(nineCharID) == 9 {
			if nineCharID[:4] == strings.ToUpper(s.Ident.FourCharacterID) {
				s.Ident.NineCharacterID = nineCharID
			}
		}
	}
	return s, nil
}

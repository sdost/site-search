package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

// Item struct to represent CSV row
type Item struct {
	Rank          int
	URL           string
	RootDomains   string
	ExternalLinks string
	MozRank       string
	MozTrust      string
}

// CheckForTerm Check website for search term.
func CheckForTerm(host string, term string) (bool, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s", host))
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	found, err := regexp.MatchString(term, string(body))
	if err != nil {
		return false, err
	}

	return found, nil
}

// ProcessCSV Unmarshall CSV data into a channel
func ProcessCSV(rc io.Reader) (ch chan Item) {
	ch = make(chan Item)
	go func() {
		r := csv.NewReader(rc)
		if _, err := r.Read(); err != nil { //read header
			log.Fatal(err)
		}
		defer close(ch)
		for {
			var rec Item
			err := UnmarshalCSV(r, &rec)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal(err)
			}
			ch <- rec
		}
	}()
	return ch
}

func main() {
	http.DefaultClient.Timeout = 10 * time.Second

	var term string
	flag.StringVar(&term, "s", "", "Search term.")
	var maxConcurrency int
	flag.IntVar(&maxConcurrency, "c", 20, "Maximum concurrency.")
	var urlFilePath string
	flag.StringVar(&urlFilePath, "i", "url.txt", "Path to URL list file.")
	var resultsFilePath string
	flag.StringVar(&resultsFilePath, "o", "results.txt", "Path to results file.")
	flag.Parse()

	csvfile, err := os.Open(urlFilePath)
	if err != nil {
		log.Fatal(err)
	}
	rows := ProcessCSV(bufio.NewReader(csvfile))

	outfile, err := os.Create(resultsFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()
	out := bufio.NewWriter(outfile)
	out.WriteString(fmt.Sprintf("Search term: %s\n", term))

	sem := make(chan bool, maxConcurrency)
	for row := range rows {
		host := row.URL
		sem <- true
		go func(host string) {
			defer func() { <-sem }()
			found, err := CheckForTerm(host, term)
			if err != nil {
				fmt.Println(err)
			}
			out.WriteString(fmt.Sprintf("%s, found: %t\n", host, found))
		}(host)
	}
	out.Flush()
}

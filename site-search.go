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
	"reflect"
	"regexp"
	"strconv"
	"time"
)

type item struct {
	Rank          int
	URL           string
	RootDomains   string
	ExternalLinks string
	MozRank       string
	MozTrust      string
}

type fieldMismatch struct {
	expected, found int
}

func (e *fieldMismatch) Error() string {
	return "CSV line fields mismatch. Expected " + strconv.Itoa(e.expected) + " found " + strconv.Itoa(e.found)
}

type unsupportedType struct {
	Type string
}

func (e *unsupportedType) Error() string {
	return "Unsupported type: " + e.Type
}

func unmarshal(reader *csv.Reader, v interface{}) error {
	record, err := reader.Read()
	if err != nil {
		return err
	}
	s := reflect.ValueOf(v).Elem()
	if s.NumField() != len(record) {
		return &fieldMismatch{s.NumField(), len(record)}
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		switch f.Type().String() {
		case "string":
			f.SetString(record[i])
		case "int":
			ival, err := strconv.ParseInt(record[i], 10, 0)
			if err != nil {
				return err
			}
			f.SetInt(ival)
		default:
			return &unsupportedType{f.Type().String()}
		}
	}
	return nil
}

func processCSV(rc io.Reader) (ch chan item) {
	ch = make(chan item)
	go func() {
		r := csv.NewReader(rc)
		if _, err := r.Read(); err != nil { //read header
			log.Fatal(err)
		}
		defer close(ch)
		for {
			var rec item
			err := unmarshal(r, &rec)
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

func checkForTerm(host string, term string) (bool, error) {
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

	csvfile, _ := os.Open(urlFilePath)
	rows := processCSV(bufio.NewReader(csvfile))

	outfile, err := os.Create(resultsFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()
	out := bufio.NewWriter(outfile)

	sem := make(chan bool, maxConcurrency)
	for row := range rows {
		host := row.URL
		sem <- true
		go func(host string) {
			defer func() { <-sem }()
			found, err := checkForTerm(host, term)
			if err != nil {
				fmt.Println(err)
			}
			out.WriteString(fmt.Sprintf("Found term [%s] in [%s]: %v\n", term, host, found))
		}(host)
	}
	out.Flush()
}

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Records []map[string]interface{}

func main() {
	if err := filepath.Walk(".", walkf); err != nil {
		panic(err)
	}
}

func parse(b []byte, target string) {
	buf := bytes.NewBuffer(b)

	r := csv.NewReader(buf)
	record, err := r.Read()
	if err != nil {
		panic(err)
	}
	title := record
	tmap := make([]func(string) interface{}, len(title))
	nmap := make([]func(string) string, len(title))
	for idx, field := range title {
		// typ may be empty string
		typ := find(field, regTyp)
		name := find(field, regName)
		fmt.Println(name)
		tmap[idx] = Wraper(typ)
		if name == "" {
			nmap[idx] = Wraper02(typ)
		} else {
			nmap[idx] = Wraper03(name)
		}
	}
	var Recs Records

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		re := map[string]interface{}{}
		for idx, field := range title {
			key := nmap[idx](field)
			re[key] = tmap[idx](record[idx])
		}
		Recs = append(Recs, re)
	}

	b, err = json.MarshalIndent(Recs, "", "    ")
	if err != nil {
		panic(err)
	}
	os.WriteFile(target, b, 0666)
}

func Wraper(typ string) func(string) interface{} {
	switch typ {
	case "{i}":
		return func(str string) interface{} { return s2i(str) }
	case "{f}":
		return func(str string) interface{} { return s2f(str) }
	default:
		return func(str string) interface{} { return str }
	}
}

func Wraper02(typ string) func(string) string {
	switch typ {
	case "{i}", "{f}":
		return func(field string) string { return strings.Replace(field, typ, "", 1) }
	default:
		return func(field string) string { return field }
	}
}

// ignore original name
func Wraper03(str string) func(string) string {
	return func(string) string { return str }
}

func s2i(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return i
}

func s2f(str string) float64 {
	f, err := strconv.ParseFloat(str, 3)
	if err != nil {
		panic(err)
	}
	return f
}

var regTyp *regexp.Regexp = regexp.MustCompile(`\{.\}`)
var regName *regexp.Regexp = regexp.MustCompile(`\[.+?\]`)

func find(str string, re *regexp.Regexp) string {
	return re.FindString(str)
}

func walkf(_path string, info fs.FileInfo, err error) error {
	if err != nil {
		panic(err)
	}
	if info.IsDir() {
		return nil
	}
	if strings.ToLower(path.Ext(_path)) == ".csv" {
		b, err := os.ReadFile(_path)
		if err != nil {
			return err
		}
		target := strings.Replace(_path, path.Ext(_path), ".json", 1)
		parse(b, target)

	}
	return nil
}

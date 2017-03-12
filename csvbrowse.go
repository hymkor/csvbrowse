package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"html"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zetamatta/go-mbcs"
)

var force_tsv = flag.Bool("t", false, "Parse as tab separated value")

func isFieldCountErr(err error) bool {
	if err == csv.ErrFieldCount {
		return true
	}
	if t := err.(*csv.ParseError); t != nil && t.Err == csv.ErrFieldCount {
		return true
	}
	return false
}

func do_file(fname string, w io.Writer) error {
	pReader, pWriter := io.Pipe()

	go func() {
		r, r_err := os.Open(fname)
		if r_err != nil {
			pWriter.CloseWithError(r_err)
			return
		}
		defer r.Close()

		scnr := bufio.NewScanner(r)
		for scnr.Scan() {
			ansi := scnr.Bytes()
			unicode, err := mbcs.AtoU(ansi)
			if err != nil {
				pWriter.CloseWithError(err)
				return
			}
			if _, err2 := fmt.Fprintln(pWriter, unicode); err2 != nil {
				pWriter.CloseWithError(err2)
				return
			}
		}
		pWriter.Close()
	}()
	defer pReader.Close()
	csvr := csv.NewReader(pReader)

	if *force_tsv || strings.HasSuffix(strings.ToLower(fname), ".tsv") {
		csvr.Comma = '\t'
	}

	tag := "th"
	for {
		cols, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if !isFieldCountErr(err) {
				return err
			}
		}
		fmt.Fprint(w, "<tr>")
		for i, c := range cols {
			fmt.Fprintf(w, `<%[1]s nowrap title="%[2]d">%[3]s</%[1]s>`,
				tag,
				i+1,
				html.EscapeString(c))
		}
		fmt.Fprintln(w, "</tr>")
		tag = "td"
	}
}

func main1(files []string, htmlpath string) error {
	w, w_err := os.Create(htmlpath)
	if w_err != nil {
		return w_err
	}
	fmt.Fprintln(w, `<html>`)
	fmt.Fprintln(w, `<head>`)
	fmt.Fprintln(w, `<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />`)
	fmt.Fprintln(w, `<style type="text/css">`)
	fmt.Fprintln(w, `table{ margin-left:1cm ; border-collapse:collapse}`)
	fmt.Fprintln(w, `table,td{ border:solid 1px gray; padding:1pt`)
	fmt.Fprintln(w, `td{ white-space:nowrap }`)
	fmt.Fprintln(w, `</style>`)
	fmt.Fprintf(w, "<title>%s</title>\n", strings.Join(files, ","))
	fmt.Fprintln(w, `</head>`)
	fmt.Fprintln(w, `<body><table border>`)
	defer func() {
		fmt.Fprintln(w, `</table></body></html>`)
		w.Close()
	}()
	for _, wildcard := range files {
		matches, matches_err := filepath.Glob(wildcard)
		if matches_err != nil {
			if err := do_file(wildcard, w); err != nil {
				return fmt.Errorf("%s: %s", wildcard, err.Error())
			}
		} else {
			for _, fname := range matches {
				if err := do_file(fname, w); err != nil {
					return fmt.Errorf("%s: %s", fname, err.Error())
				}
			}
		}
	}
	return nil
}

const htmlname = "tmp.html"

func main() {
	flag.Parse()
	htmlpath := filepath.Join(os.Getenv("TEMP"), htmlname)
	if err := main1(flag.Args(), htmlpath); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	cmd1 := exec.Cmd{
		Path: "cmd.exe",
		Args: []string{"/c", "start", htmlpath},
	}
	cmd1.Run()
}

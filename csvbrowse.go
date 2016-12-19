package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zetamatta/go-mbcs"
)

var force_tsv = flag.Bool("t", false, "Parse as tab separated value")

func do_file(fname string, w io.Writer) error {
	r, r_err := os.Open(fname)
	if r_err != nil {
		return r_err
	}
	defer r.Close()

	ansi_all, ansi_all_err := ioutil.ReadAll(r)
	if ansi_all_err != nil {
		return ansi_all_err
	}

	unicode_all, unicode_all_err := mbcs.AtoU(ansi_all)
	if unicode_all_err != nil {
		return unicode_all_err
	}
	csvr := csv.NewReader(strings.NewReader(unicode_all))

	if *force_tsv || strings.HasSuffix(strings.ToLower(fname), ".tsv") {
		csvr.Comma = '\t'
	}

	for {
		cols, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if err == csv.ErrFieldCount {
				goto safe
			}
			if t := err.(*csv.ParseError); t != nil && t.Err == csv.ErrFieldCount {
				goto safe
			}
			return err
		}
	safe:
		fmt.Fprint(w, "<tr>")
		for i, c := range cols {
			fmt.Fprintf(w, `<td nowrap title="%d">%s</td>`, i+1, html.EscapeString(c))
		}
		fmt.Fprintln(w, "</tr>")
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
	fmt.Fprintln(w, `</head>`)
	fmt.Fprintf(w, "<title>%s</title>\n", strings.Join(files, ","))
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

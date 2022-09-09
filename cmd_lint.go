package shodo

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
)

func doLint(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	fs := flag.NewFlagSet("shodo lint", flag.ContinueOnError)
	fs.SetOutput(errStream)
	if err := fs.Parse(argv); err != nil {
		return err
	}
	files := fs.Args()
	if len(files) < 1 {
		return errors.New("no files specified")
	}

	body, err := os.ReadFile(files[0])
	if err != nil {
		return err
	}

	c, err := newClient()
	if err != nil {
		return err
	}
	id, err := c.CreateLint(ctx, string(body))
	if err != nil {
		return err
	}
	log.Println("Linting...")
	r, err := c.LintResult(ctx, id)
	if err != nil {
		return err
	}

	runes := []rune(string(body))
	substr := func(rs []rune, from, to int) string {
		if from < 0 {
			from = 0
		}
		if to > len(rs) {
			to = len(rs)
		}
		return string(rs[from:to])
	}

	for _, m := range r.Messages {
		co := color.New(color.FgRed)
		if m.Severity == severityWraning {
			co = color.New(color.FgYellow)
		}
		buf := &bytes.Buffer{}
		fix := ""
		if m.After != "" {
			fix = fmt.Sprintf("（→ %s）", m.After)
		}
		buf.WriteString(substr(runes, m.Index-10, m.Index))
		co.Fprint(buf, substr(runes, m.Index, m.IndexTo)+fix)
		buf.WriteString(substr(runes, m.IndexTo, m.IndexTo+10))
		fmt.Fprintln(outStream, m.From.String(), m.Message)
		fmt.Fprintln(outStream, "    "+strings.Replace(buf.String(), "\n", " ", -1))
	}
	return nil
}

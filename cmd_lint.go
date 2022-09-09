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
	"sync"

	"github.com/fatih/color"
	"golang.org/x/sync/errgroup"
)

type Lint struct {
	Result        *LintResult
	File, Content string
}

type Lints struct {
	l  []*Lint
	mu sync.RWMutex
}

func (ls *Lints) Append(l *Lint) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.l = append(ls.l, l)
}

func (ls *Lints) List() []*Lint {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.l
}

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

	log.Println("Linting...")
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(6)
	ls := &Lints{}
	for _, f := range files {
		f := f
		g.Go(func() error {
			b, err := os.ReadFile(f)
			if err != nil {
				return err
			}
			body := string(b)

			c, err := newClient()
			if err != nil {
				return err
			}
			id, err := c.CreateLint(ctx, body)
			if err != nil {
				return err
			}
			r, err := c.LintResult(ctx, id)
			if err != nil {
				return err
			}
			ls.Append(&Lint{
				Result:  r,
				File:    f,
				Content: body,
			})
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}
	lints := ls.List()

	single := len(lints) == 1
	for _, l := range lints {
		if !single {
			fmt.Fprintf(outStream, "%s:\n", l.File)
		}
		runes := []rune(l.Content)
		substr := func(rs []rune, from, to int) string {
			if from < 0 {
				from = 0
			}
			if to > len(rs) {
				to = len(rs)
			}
			return string(rs[from:to])
		}

		for _, m := range l.Result.Messages {
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
	}
	return nil
}

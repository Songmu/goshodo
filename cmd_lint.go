package goshodo

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

	lines []string
}

// rune range
func (l *Lint) Correspond(m LintMessage, bound int) (string, string, string) {
	return substr(l.Content, m.Index-bound, m.Index),
		substr(l.Content, m.Index, m.IndexTo),
		substr(l.Content, m.IndexTo, m.IndexTo+bound)
}

// 0 origin
func (l *Lint) Line(n int) string {
	if l.lines == nil {
		l.lines = strings.Split(l.Content, "\n")
	}
	if n < 0 || len(l.lines) <= n {
		return ""
	}
	return l.lines[n]
}

func substr(content string, from, to int) string {
	rs := []rune(content)
	if from < 0 {
		from = 0
	}
	if to > len(rs) {
		to = len(rs)
	}
	return string(rs[from:to])
}

type lints struct {
	l  []*Lint
	mu sync.RWMutex
}

func (ls *lints) Append(l *Lint) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.l = append(ls.l, l)
}

func (ls *lints) List() []*Lint {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.l
}

func doLint(ctx context.Context, argv []string, outStream, errStream io.Writer) error {
	fs := flag.NewFlagSet("goshodo lint", flag.ContinueOnError)
	fs.SetOutput(errStream)
	format := fs.String("f", "", "format")
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
	ls := &lints{}
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
	if *format == "checkstyle" {
		return fprintCheckStyle(outStream, lints)
	}

	single := len(lints) == 1
	for _, l := range lints {
		if !single {
			fmt.Fprintf(outStream, "%s:\n", l.File)
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
			pre, match, post := l.Correspond(m, 10)
			buf.WriteString(pre)
			co.Fprint(buf, match+fix)
			buf.WriteString(post)
			fmt.Fprintln(outStream, m.From.String(), m.Message)
			fmt.Fprintln(outStream, "    "+strings.Replace(buf.String(), "\n", " ", -1))
		}
	}
	return nil
}

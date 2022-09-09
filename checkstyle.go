package goshodo

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type CheckStyle struct {
	XMLName xml.Name         `xml:"checkstyle"`
	Version string           `xml:"version,attr"`
	Files   []CheckStyleFile `xml:"file"`
}

type CheckStyleFile struct {
	XMLName xml.Name          `xml:"file"`
	Name    string            `xml:"name,attr"`
	Errors  []CheckStyleError `xml:"error"`
}

type CheckStyleError struct {
	XMLName  xml.Name `xml:"error"`
	Line     int      `xml:"line,attr"`
	Column   int      `xml:"column,attr,omitempty"`
	Severity string   `xml:"severity,attr,omitempty"`
	Message  string   `xml:"message,attr"`
	Source   string   `xml:"source,attr"`
}

func fprintCheckStyle(w io.Writer, lints []*Lint) error {
	var c CheckStyle
	for _, l := range lints {
		f := CheckStyleFile{
			Name: l.File,
		}
		for _, m := range l.Result.Messages {
			after := m.After
			if after == "トル" {
				after = ""
			}
			line := l.Line(m.From.Line)
			lineLen := len([]rune(line))

			pre, match, post := l.Correspond(m, lineLen)
			var postMsg string
			if !strings.Contains(match, "\n") {
				pres := strings.Split(pre, "\n")
				p := pres[len(pres)-1]
				po := strings.Split(post, "\n")[0]
				postMsg = fmt.Sprintf("```suggestsion\n%s%s%s\n```", p, after, po)
			} else {
				fix := ""
				if m.After != "" {
					fix = fmt.Sprintf("（→ %s）", m.After)
				}
				pre, match, post := l.Correspond(m, 10)
				postMsg = fmt.Sprintf("    %s~~%s~~%s%s", pre, match, fix, post)
			}
			msg := fmt.Sprintf("%s\n%s", m.Message, postMsg)
			f.Errors = append(f.Errors, CheckStyleError{
				Line:     m.From.Line + 1,
				Column:   m.From.Ch + 1,
				Severity: string(m.Severity),
				Message:  msg,
			})
		}
		c.Files = append(c.Files, f)
	}
	if _, err := fmt.Fprint(w, xml.Header); err != nil {
		return err
	}
	return xml.NewEncoder(w).Encode(c)
}

package goshodo

import (
	"encoding/xml"
	"fmt"
	"io"
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
			fix := ""
			if m.After != "" {
				fix = fmt.Sprintf("（→ %s）", m.After)
			}
			pre, match, post := l.Correspond(m)
			msg := fmt.Sprintf("%s\n    %s~~%s~~%s%s", m.Message, pre, match, fix, post)
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

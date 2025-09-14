package db

import (
	"fmt"
	"html/template"
	"strings"
	"text/scanner"
)

// mark is a rudimentary markup. Tags:
//
// |<color>| where <color> is an HTML color (css notation)
// {<note>} adds a <note> associated with the current line
//
// May add more schtuff later. For now, parse each line into separate table row with notes in adjacent col

type MarkText struct {
	TextLines []string
}

func NewMarkText(raw string) *MarkText {
	i := 0
	tl := []string{}
	for l := range strings.Lines(raw) {
		//fmt.Println(i, l)
		tl = append(tl, l)
		i++
	}
	return &MarkText{TextLines: tl}
}

const (
	Text      string = "TEXT"
	Color     string = "COLOR"
	Note      string = "NOTE"
	ColorText string = "COLORTEXT"
)

// TODO: make this configurable?
type Row struct {
	Text  string
	Notes string
}
type Col struct {
	Row []Row
}

const ROWS = 30

func (m *MarkText) PrettyText(rows_per_col int) template.HTML {
	s := scanner.Scanner{}
	var b strings.Builder
	cols := []Col{}
	state := Text
	//last_state := TEXT
	b.WriteString("<table class='lyrics'>")
	cidx := 0
	//cols = append(cols, Col{Text: []string{}, Notes: []string{}})
	cols = append(cols, Col{[]Row{}})
	if rows_per_col == 0 {
		rows_per_col = (len(m.TextLines) + 1) / 2
	}
	for i, l := range m.TextLines {
		// search for tags
		s.Init(strings.NewReader(l))
		s.Mode ^= scanner.ScanChars | scanner.ScanStrings
		//fmt.Println("line: ", i)
		open_span := false
		//b.WriteString("<tr><td class='lyrics'>")
		//cols[cidx].Text = append(cols[cidx].Text, "<tr><td class='lyrics'>")
		//ridx := len(cols[cidx].Text)
		row_text := "<td class='lyrics'>"
		note := ""

		for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			switch tok {
			case '|':
				switch state {
				case ColorText:
					if open_span {
						//b.WriteString("</span>")
						row_text += "</span>"
						open_span = false
					}
					fallthrough
				case Text:
					state = Color
				case Color:
					state = ColorText
				}
				continue
			case '{':
				state = Note
				continue
			case '}':
				state = Text
				continue
			}

			//fmt.Println(state)
			switch state {
			case Text:
				fallthrough
			case ColorText:
				//b.WriteString(s.TokenText())
				row_text += s.TokenText()
				if s.Peek() == ' ' {
					//b.WriteRune(' ')
					row_text += " "
				}
			case Note:
				note += s.TokenText()
				if s.Peek() == ' ' {
					note += " "
				}
			case Color:
				//b.WriteString("<span style='color:")
				//b.WriteString(s.TokenText())
				//b.WriteString(";'>")
				row_text += fmt.Sprintf("<span class='%s'>", s.TokenText())
				open_span = true
			}
			//fmt.Printf("%s %s: %s\n", state, s.Position, s.TokenText())
		}
		if open_span {
			//b.WriteString("</span>")
			row_text += "<span>"
			open_span = false
		}
		//b.WriteString("</td>")
		row_text += "&nbsp;</td>\n"
		note = fmt.Sprintf("<td class='note'>%s</td>\n", note)
		//b.WriteString(fmt.Sprintf("<td class='note'>%s</td></tr>", note))
		cols[cidx].Row = append(cols[cidx].Row, Row{row_text, note})
		//cols[cidx].Notes = append(cols[cidx].Notes, note)
		if i%rows_per_col == (rows_per_col - 1) {
			//fmt.Println("new col")
			cols = append(cols, Col{[]Row{}})
			cidx += 1
		}
	}
	//fmt.Println("Number of columns: ", len(cols))
	max_rows := len(cols[0].Row)
	for r := 0; r < max_rows; r++ {
		b.WriteString("<tr>")
		for c := 0; c < len(cols); c++ {
			if r < len(cols[c].Row) {
				b.WriteString(cols[c].Row[r].Text)
				b.WriteString(cols[c].Row[r].Notes)
			}
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</table>")
	return template.HTML(b.String())
}

package lolregi

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// ParseLegalValues parses enum values from description.
//
// returns nil if "Legal values: " is not found.
func ParseLegalValues(desc string) []string {
	const prefix = "Legal values:"

	lvIdx := strings.Index(desc, prefix)
	if lvIdx == -1 {
		return nil
	}

	s := strings.TrimSpace(desc[lvIdx+len(prefix):])
	vals := strings.Split(s, ",")

	var ret []string
	for i, v := range vals {
		v = strings.TrimSpace(v)

		if i == len(vals)-1 {
			v = strings.TrimSuffix(v, ")")
		}

		ret = append(ret, v)
	}
	return ret
}

// IsEmpty returns true if the element is empty.
func IsEmpty(s *goquery.Selection) bool {
	//TODO
	if s.Text() != "" {
		return false
	}

	if s.Children().Length() == 0 {
		return true
	}
	return false
}

// Usage: in godebug, `print htmlOf(sel)`
func htmlOf(s *goquery.Selection) {
	html, err := s.Html()
	if err != nil {
		panic(err)
	}
	html = gohtml.Format(html)
	if err := ioutil.WriteFile("./dev.html", []byte(html), os.ModePerm); err != nil {
		panic(err)
	}
}

// NewDocument creates a new html document
// and removes some useless stuffs to make debugging easier.
func NewDocument() *goquery.Document {
	pageDoc, err := goquery.NewDocument("https://developer.riotgames.com/api/methods")
	if err != nil {
		panic(err)
	}

	doc := goquery.NewDocumentFromNode(pageDoc.Find(`#resources`).Nodes[0])
	doc.Find(`table`).RemoveClass(`table`)

	doc.Find(`.response, .sandbox_header`).Remove()
	doc.Find(`.resource > .heading`).Remove()
	all := doc.Find(`*`).RemoveAttr(`style`).RemoveAttr(`onclick`)

	for i, n := range all.Nodes {
		s := all.Eq(i)
		if isUselessNode(n, s) {
			s.Remove()
		}
	}

	return doc
}

func isUselessNode(n *html.Node, s *goquery.Selection) bool {
	switch n.DataAtom {
	case atom.Input:
		// hidden input field 'method_id'
		if name, ok := s.Attr("name"); ok && name == "method_id" {
			return true
		}

	case atom.Div:
		if id, ok := s.Attr("id"); ok && id == "inputs-link" {
			return true
		}

	case atom.A:
		if href, ok := s.Attr("href"); ok && strings.HasPrefix(href, "#!/") {
			s.RemoveAttr("href")
		}
	}

	return false
}

func assert(s *goquery.Selection, selector string) {
	if len(s.Nodes) == 0 {
		panic(fmt.Sprintf("Assertion failed for selector '%s', because len == 0", selector))
	}

	if !s.Is(selector) {
		var (
			html string
			err  error
		)
		if s.Parent() == nil {
			html, err = s.Html()
		} else {
			html, err = s.Parent().Html()
		}

		if err != nil {
			panic(err)
		}
		panic(fmt.Sprintf("Assertion failed for selector '%s'\nNodes: %v,\n----- HTML -----\n%s", selector, s.Nodes, html))
	}
}

// parameterName converts strings like 'teamIds', 'platformId' to 'teamIDs', 'platformID'
func parameterName(name string) string {
	if strings.HasSuffix(name, "Ids") {
		return strings.TrimSuffix(name, "Ids") + "IDs"
	}

	return lintName(name)
}

// See: https://github.com/golang/lint/blob/master/lint.go
// commonInitialisms is a set of common initialisms.
// Only add entries that are highly unlikely to be non-initialisms.
// For instance, "ID" is fine (Freudian code is rare), but "AND" is not.
var commonInitialisms = map[string]bool{
	"API":   true,
	"ASCII": true,
	"CPU":   true,
	"CSS":   true,
	"DNS":   true,
	"EOF":   true,
	"GUID":  true,
	"HTML":  true,
	"HTTP":  true,
	"HTTPS": true,
	"ID":    true,
	"IP":    true,
	"JSON":  true,
	"LHS":   true,
	"QPS":   true,
	"RAM":   true,
	"RHS":   true,
	"RPC":   true,
	"SLA":   true,
	"SMTP":  true,
	"SQL":   true,
	"SSH":   true,
	"TCP":   true,
	"TLS":   true,
	"TTL":   true,
	"UDP":   true,
	"UI":    true,
	"UID":   true,
	"UUID":  true,
	"URI":   true,
	"URL":   true,
	"UTF8":  true,
	"VM":    true,
	"XML":   true,
	"XSRF":  true,
	"XSS":   true,
}

// lintName returns a different name if it should be different.
func lintName(name string) (should string) {
	// Fast path for simple cases: "_" and all lowercase.
	if name == "_" {
		return name
	}
	allLower := true
	for _, r := range name {
		if !unicode.IsLower(r) {
			allLower = false
			break
		}
	}
	if allLower {
		return name
	}

	// Split camelCase at any lower->upper transition, and split on underscores.
	// Check each word for common initialisms.
	runes := []rune(name)
	w, i := 0, 0 // index of start of word, scan
	for i+1 <= len(runes) {
		eow := false // whether we hit the end of a word
		if i+1 == len(runes) {
			eow = true
		} else if runes[i+1] == '_' {
			// underscore; shift the remainder forward over any run of underscores
			eow = true
			n := 1
			for i+n+1 < len(runes) && runes[i+n+1] == '_' {
				n++
			}

			// Leave at most one underscore if the underscore is between two digits
			if i+n+1 < len(runes) && unicode.IsDigit(runes[i]) && unicode.IsDigit(runes[i+n+1]) {
				n--
			}

			copy(runes[i+1:], runes[i+n+1:])
			runes = runes[:len(runes)-n]
		} else if unicode.IsLower(runes[i]) && !unicode.IsLower(runes[i+1]) {
			// lower->non-lower
			eow = true
		}
		i++
		if !eow {
			continue
		}

		// [w,i) is a word.
		word := string(runes[w:i])
		if u := strings.ToUpper(word); commonInitialisms[u] {
			// Keep consistent case, which is lowercase only at the start.
			if w == 0 && unicode.IsLower(runes[w]) {
				u = strings.ToLower(u)
			}
			// All the common initialisms are ASCII,
			// so we can replace the bytes exactly.
			copy(runes[w:], []rune(u))
		} else if w > 0 && strings.ToLower(word) == word {
			// already all lowercase, and not the first word, so uppercase the first character.
			runes[w] = unicode.ToUpper(runes[w])
		}
		w = i
	}
	return string(runes)
}

package lolregi

import (
	"bytes"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func TestParseResponseField(t *testing.T) {
	type testData struct {
		HTML, Name, RawType, Desc string
	}

	datas := []testData{
		{
			`<tr>
		<td>mapId</td>
		<td>int</td>
		<td>Match map ID</td>
		</tr>`,
			"mapId", "int", "Match map ID",
		},
		{
			`<tr>
		<td>matchCreation</td>
		<td>long</td>
		<td>Match creation time. Designates when the team select lobby is created and/or the match is made through match making, not when the game actually starts.</td>
		</tr>`,
			"matchCreation",
			"long",
			"Match creation time. Designates when the team select lobby is created and/or the match is made through match making, not when the game actually starts.",
		},
		{
			`<tr>
			<td>matchDuration</td>
			<td>long</td>
			<td>Match duration</td>
			</tr>`,
			"matchDuration", "long", "Match duration",
		},
	}

	for _, d := range datas {
		d.HTML = "<div><table><tbody>" + d.HTML + "</tbody></table></div>" // IMPORTANT!
		root, err := html.Parse(bytes.NewReader([]byte(d.HTML)))
		if err != nil {
			t.Fatal(err)
			return
		}
		var buf bytes.Buffer
		if err := html.Render(&buf, root); err != nil {
			t.Fatal(err)
			return
		}
		t.Logf(`Tag: "%v"`, buf.String())

		doc := goquery.NewDocumentFromNode(root)

		t.Logf("Expected name=%s, rawType=%s, desc=%s", d.Name, d.RawType, d.Desc)

		name, rawType, desc := parseField(doc.Find("tr"))
		if name != d.Name || rawType != d.RawType || desc != d.Desc {
			t.Fatalf("Invalid name=%s, rawType=%s, desc=%s", name, rawType, desc)
		}
	}
}

func TestParseResponseErrors(t *testing.T) {
	const html = `<div class="api_block">
	<h4>
	Response Errors
	</h4>
	<table>
	<tbody>
	<tr class="odd">
	<td>
	400
	</td>
	<td>
	Bad request
	</td>
	</tr>
	<tr class="odd">
	<td>
	401
	</td>
	<td>
	Unauthorized
	</td>
	</tr>
	<tr class="odd">
	<td>
	429
	</td>
	<td>
	Rate limit exceeded
	</td>
	</tr>
	<tr class="odd">
	<td>
	500
	</td>
	<td>
	Internal server error
	</td>
	</tr>
	<tr class="odd">
	<td>
	503
	</td>
	<td>
	Service unavailable
	</td>
	</tr>
	</tbody>
	</table>
	</div>`
	expected := []ResponseError{
		{400, "Bad request"},
		{401, "Unauthorized"},
		{429, "Rate limit exceeded"},
		{500, "Internal server error"},
		{503, "Service unavailable"},
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(html)))
	if err != nil {
		t.Fatal(err)
		return
	}

	errors := parseResponseErrors(doc.Find("div.api_block"))
	for i, e := range errors {

		if expected[i] != e {
			t.Fatal(`Expected: `, expected[i], e)
			return
		}
	}
}

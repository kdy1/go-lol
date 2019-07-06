package lol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SpellRange struct {
	Self   bool
	Ranges []int32
}

// UnmarshalJSON handles strange value of the spell range.
func (sr *SpellRange) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == `"self"` {
		sr.Self = true
		return nil
	}
	sr.Self = false

	// Some resources (and some versions) returns an array.
	if strings.HasPrefix(str, "[") {
		sr.Ranges = make([]int32, 0)
		return json.Unmarshal(data, &sr.Ranges)
	}

	type spellRange struct {
		Ranges []int32 `json:"Ranges"`
	}

	r := new(spellRange)
	if err := json.Unmarshal(data, r); err != nil {
		return err
	}
	sr.Ranges = r.Ranges
	return nil
}

// concat ids with ','
func joinIDs(ids []int64) string {
	var buf bytes.Buffer

	for i, id := range ids {
		buf.WriteString(strconv.FormatInt(id, 10))
		if len(ids) != i+1 { // if it's not last one, write ','
			buf.WriteRune(',')
		}
	}

	return buf.String()
}

// closeBody is used to close res.Body.
// Prior to calling Close, it also tries to Read a small amount to see an EOF.
// Not seeing an EOF can prevent HTTP Transports from reusing connections.
//
// See: https://godoc.org/google.golang.org/api/googleapi#CloseBody
func closeBody(res *http.Response) {
	if res == nil || res.Body == nil {
		return
	}
	// Justification for 3 byte reads: two for up to "\r\n" after
	// a JSON/XML document, and then 1 to see EOF if we haven't yet.
	// TODO(bradfitz): detect Go 1.3+ and skip these reads.
	// See https://codereview.appspot.com/58240043
	// and https://codereview.appspot.com/49570044
	buf := make([]byte, 1)
	for i := 0; i < 3; i++ {
		_, err := res.Body.Read(buf)
		if err != nil {
			break
		}
	}
	res.Body.Close()
}

func convertToString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case []int64:
		return joinIDs(v)
	case []string:
		var buf bytes.Buffer

		for i, s := range v {
			buf.WriteString(s)
			if len(v) != i+1 { // if it's not last one, write ','
				buf.WriteRune(',')
			}
		}

		return buf.String()
	default:
		return fmt.Sprint(v)
	}
}

func ParseEpochMilliseconds(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

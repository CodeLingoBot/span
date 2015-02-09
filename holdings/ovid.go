// Package holdings contains wrappers for various holding file formats.
package holdings

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// DelayPattern is how moving walls are expressed in OVID format
var DelayPattern = regexp.MustCompile(`^-(\d+)(M|Y)$`)

// Holding contains a single holding
type Holding struct {
	EZBID        int           `xml:"ezb_id,attr"`
	Title        string        `xml:"title"`
	Publishers   string        `xml:"publishers"`
	PISSN        []string      `xml:"EZBIssns>p-issn"`
	EISSN        []string      `xml:"EZBIssns>e-issn"`
	Entitlements []Entitlement `xml:"entitlements>entitlement"`
}

// Entitlement holds a single OVID entitlement
type Entitlement struct {
	Status     string `xml:"status,attr"`
	URL        string `xml:"url"`
	Anchor     string `xml:"anchor"`
	FromYear   int    `xml:"begin>year"`
	FromVolume int    `xml:"begin>volume"`
	FromIssue  int    `xml:"begin>issue"`
	FromDelay  string `xml:"begin>delay"`
	ToYear     int    `xml:"end>year"`
	ToVolume   int    `xml:"end>volume"`
	ToIssue    int    `xml:"end>issue"`
	ToDelay    string `xml:"end>delay"`
}

// String returns a string representation of an Entitlement
func (e *Entitlement) String() string {
	delay, _ := e.Delay()
	unescaped, _ := url.QueryUnescape(e.URL)
	boundary, _ := e.Boundary()
	return fmt.Sprintf("<Entitlement status=%s url=%s range=%d/%d/%d-%d/%d/%d boundary=%s delay=%0.2f>",
		e.Status, unescaped, e.FromYear, e.FromVolume, e.FromIssue, e.ToYear, e.ToVolume, e.ToIssue,
		boundary, delay.Hours())
}

// IssnHolding maps an ISSN to a holdings.Holding struct
type IssnHolding map[string]Holding

// IsilIssnHolding maps an ISIL to an IssnHolding map
type IsilIssnHolding map[string]IssnHolding

// Isils returns available ISILs in this IsilIssnHolding map
func (iih *IsilIssnHolding) Isils() []string {
	var keys []string
	for k, _ := range *iih {
		keys = append(keys, k)
	}
	return keys
}

// ParseDelay parses delay strings like '-1M', '-3Y', ... into a time.Duration
func ParseDelay(s string) (d time.Duration, err error) {
	ms := DelayPattern.FindStringSubmatch(s)
	if len(ms) == 3 {
		value, err := strconv.Atoi(ms[1])
		if err != nil {
			return d, err
		}
		switch {
		case ms[2] == "Y":
			d, err = time.ParseDuration(fmt.Sprintf("-%dh", value*8760))
		case ms[2] == "M":
			d, err = time.ParseDuration(fmt.Sprintf("-%dh", value*720))
		default:
			return d, fmt.Errorf("unknown unit: %s", ms[2])
		}
	} else {
		return d, fmt.Errorf("unknown format: %s", s)
	}
	return d, nil
}

// Delay returns the specified delay as `time.Duration`
func (e *Entitlement) Delay() (d time.Duration, err error) {
	if e.FromDelay != "" {
		return ParseDelay(e.FromDelay)
	}
	if e.ToDelay != "" {
		return ParseDelay(e.ToDelay)
	}
	return d, nil
}

// Boundary returns the last date before the moving wall restriction becomes effective
func (e *Entitlement) Boundary() (d time.Time, err error) {
	delay, err := e.Delay()
	if err != nil {
		return d, err
	}
	return time.Now().Add(delay), nil
}

// HoldingsMap creates an ISSN[Holding] struct from a reader
func HoldingsMap(reader io.Reader) (h IssnHolding) {
	h = make(map[string]Holding)
	decoder := xml.NewDecoder(reader)
	var tag string
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			tag = se.Name.Local
			if tag == "holding" {
				var item Holding
				decoder.DecodeElement(&item, &se)
				for _, id := range item.EISSN {
					h[id] = item
				}
				for _, id := range item.PISSN {
					h[id] = item
				}
			}
		}
	}
	return h
}

package holdings

import (
	"io"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseDelay(t *testing.T) {
	var tests = []struct {
		s   string
		d   time.Duration
		err error
	}{
		{"-0M", time.Duration(0), nil},
		{"-1M", time.Duration(-1) * month, nil},
		{"-2M", time.Duration(-2) * month, nil},
		{"-1Y", time.Duration(-1) * year, nil},
		{"-1D", time.Duration(0), errUnknownFormat},
		{"-1", time.Duration(0), errUnknownFormat},
		{"129", time.Duration(0), errUnknownFormat},
		{"AB", time.Duration(0), errUnknownFormat},
		{"-111m", time.Duration(0), errUnknownFormat},
		{"0.1M", time.Duration(0), errUnknownFormat},
	}

	for _, tt := range tests {
		d, err := ParseDelay(tt.s)
		if d != tt.d {
			t.Errorf("ParseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
		}
		if err != nil {
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("ParseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
				}
			} else {
				t.Errorf("ParseDelay(%s) => %v, %v, want %v, %v", tt.s, d, err, tt.d, tt.err)
			}
		}
	}
}

func TestDelay(t *testing.T) {
	var tests = []struct {
		e   Entitlement
		d   time.Duration
		err error
	}{
		{Entitlement{FromDelay: "-1M"}, time.Duration(-2592000000000000), nil},
		{Entitlement{ToDelay: "-1M"}, time.Duration(-2592000000000000), nil},
		{Entitlement{FromDelay: "-1M", ToDelay: "-1M"}, time.Duration(-2592000000000000), nil},
		{Entitlement{FromDelay: "-1M", ToDelay: "-2M"}, time.Duration(0), errDelayMismatch},
		{Entitlement{FromDelay: "-2M", ToDelay: "-1M"}, time.Duration(0), errDelayMismatch},
	}
	for _, tt := range tests {
		d, err := tt.e.Delay()
		if d != tt.d {
			t.Errorf("e.Delay() => %v, %v, want %v, %v", d, err, tt.d, tt.err)
		}
		if err != nil {
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("e.Delay() => %v, %v, want %v, %v", d, err, tt.d, tt.err)
				}
			} else {
				t.Errorf("e.Delay() => %v, %v, want %v, %v", d, err, tt.d, tt.err)
			}
		}
	}
}

func TestBoundary(t *testing.T) {
	margin := 100 * time.Microsecond
	var tests = []struct {
		e   Entitlement
		d   time.Time
		err error
	}{
		{Entitlement{FromDelay: "-0M"}, time.Now(), nil},
		{Entitlement{FromDelay: "0M"}, time.Now(), errUnknownFormat},
	}
	for _, tt := range tests {
		d, err := tt.e.Boundary()
		if err != nil {
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("e.Boundary() => %v, %v, want %v, %v", d, err, tt.d, tt.err)
				}
			} else {
				t.Errorf("e.Boundary() => %v, %v, want %v, %v", d, err, tt.d, tt.err)
			}
		}
		if d.Sub(tt.d) > time.Duration(margin) {
			t.Errorf("e.Boundary() => %v, %v, want %v, %v", d, err, tt.d, tt.err)
		}
	}
}

func TestHoldingsMap(t *testing.T) {
	var tests = []struct {
		r io.Reader
		m map[string]Holding
	}{
		{strings.NewReader(`
<holding ezb_id = "1">
  <title><![CDATA[Journal of Molecular Modeling]]></title>
  <publishers><![CDATA[Springer]]></publishers>
  <EZBIssns>
    <p-issn>1610-2940</p-issn>
    <e-issn>0948-5023</e-issn>
  </EZBIssns>
  <entitlements>
    <entitlement status = "subscribed">
      <url>http%3A%2F%2Flink.springer.com%2Fjournal%2F894</url>
      <anchor>natli_springer</anchor>
      <begin>
        <year>1995</year>
        <volume>1</volume>
      </begin>
      <end>
        <year>2002</year>
        <volume>8</volume>
      </end>
      <available><![CDATA[Nationallizenz]]></available>
    </entitlement>
    <entitlement status = "subscribed">
      <url>http%3A%2F%2Flink.springer.com%2Fjournal%2F894</url>
      <anchor>springer</anchor>
      <available><![CDATA[Konsortiallizenz - Gesamter Zeitraum]]></available>
    </entitlement>
  </entitlements>
</holding>`), map[string]Holding{
			"1610-2940": {
				EZBID:      1,
				Title:      "Journal of Molecular Modeling",
				Publishers: "Springer",
				PISSN:      []string{"1610-2940"},
				EISSN:      []string{"0948-5023"},
				Entitlements: []Entitlement{
					{
						Status:     "subscribed",
						URL:        "http://link.springer.com/journal/894",
						Anchor:     "natli_springer",
						FromYear:   1995,
						FromVolume: 1,
						ToYear:     2002,
						ToVolume:   8,
					},
					{
						Status: "subscribed",
						URL:    "http://link.springer.com/journal/894",
						Anchor: "springer",
					},
				},
			},
		}},
	}
	for _, tt := range tests {
		m := HoldingsMap(tt.r)
		if reflect.DeepEqual(m, tt.m) {
			t.Errorf("HoldingsMap(%v) => %+v, want %+v", tt.r, m, tt.m)
		}
	}
}

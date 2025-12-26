package adif

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	input := `
<call:5>K1ABC <qso_date:8>20230101 <mode:2>CW <eor>
<call:6>JA1XYZ <qso_date:8>20230102 <mode:3>SSB <qsl_rcvd:1>Y <eor>
`
	r := strings.NewReader(input)
	records, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	if records[0]["CALL"] != "K1ABC" {
		t.Errorf("Record 1 CALL mismatch: %v", records[0]["CALL"])
	}
	if records[1]["MODE"] != "SSB" {
		t.Errorf("Record 2 MODE mismatch: %v", records[1]["MODE"])
	}
	if records[1]["QSL_RCVD"] != "Y" {
		t.Errorf("Record 2 QSL_RCVD mismatch: %v", records[1]["QSL_RCVD"])
	}
}

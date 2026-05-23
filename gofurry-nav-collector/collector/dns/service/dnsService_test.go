package service

import (
	"net"
	"testing"
)

func TestLookupGeoASNNilReadersFallsBackToUnknown(t *testing.T) {
	country, city, asn, isp := lookupGeoASN(net.ParseIP("203.0.113.10"), nil, nil, nil)

	if country != "Unknown" || city != "Unknown" || asn != "Unknown" || isp != "Unknown" {
		t.Fatalf("lookupGeoASN() = (%q, %q, %q, %q), want all Unknown", country, city, asn, isp)
	}
}

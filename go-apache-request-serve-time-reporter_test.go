package main

import (
	"testing"
)

func TestReverseHostname(t *testing.T) {
	const hostname_in, hostname_out = "test.a.is.this", "a.test"

	if x := ReverseHostname(hostname_in); x != hostname_out {
		t.Errorf("ReverseHostname(%v) = %v, want %v", hostname_in, x, hostname_out)
	}
}

func TestMetricPrefix(t *testing.T) {
	const prefix_in, prefix_out = "test", "test.apache"

	if x := MetricPrefix(prefix_in); x != prefix_out {
		t.Errorf("MetricPrefix(%v) = %v, want %v", prefix_in, x, prefix_out)
	}
}

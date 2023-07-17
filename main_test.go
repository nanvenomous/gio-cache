package main

import "testing"

func stringsEqual(t *testing.T, expected, actual string) {
	if actual != expected {
		t.Errorf("Expected: %s, Actual:%s", expected, actual)
	}
}

func assertCorrectDirective(t *testing.T, acceptEncoding, expectedDirective string) {
	var bin = getMatchingBinary(acceptEncoding)
	stringsEqual(t, expectedDirective, bin.Directive)
}

func TestGetMatchingBinary(t *testing.T) {
	assertCorrectDirective(t, "br;q=1.0, gzip;q=0.8, *;q=0.1", "identity")
	binaries = []*Binary{
		{Path: brotliWASMPath, Directive: "br", Valid: true},
		{Path: zstdWASMPath, Directive: "zstd"},
		{Path: gzipWASMPath, Directive: "gzip"},
		{Path: commonWASMPath, Directive: "identity"},
	}
	assertCorrectDirective(t, "br;q=1.0, gzip;q=0.8, *;q=0.1", "br")
	binaries = []*Binary{
		{Path: brotliWASMPath, Directive: "br"},
		{Path: zstdWASMPath, Directive: "zstd"},
		{Path: gzipWASMPath, Directive: "gzip", Valid: true},
		{Path: commonWASMPath, Directive: "identity"},
	}
	assertCorrectDirective(t, "br;q=1.0, gzip;q=0.8, *;q=0.1", "gzip")
	binaries = []*Binary{
		{Path: brotliWASMPath, Directive: "br", Valid: true},
		{Path: zstdWASMPath, Directive: "zstd"},
		{Path: gzipWASMPath, Directive: "gzip", Valid: true},
		{Path: commonWASMPath, Directive: "identity"},
	}
	assertCorrectDirective(t, "gzip;q=1.0, br;q=0.8, *;q=0.1", "br")
}

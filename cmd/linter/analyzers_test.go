package main

import (
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzersAreValid(t *testing.T) {
	if err := analysis.Validate([]*analysis.Analyzer{PanicAnalyzer, LogFatalAnalyzer, OsExitAnalyzer}); err != nil {
		t.Fatal(err)
	}
}

func TestPanicAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), PanicAnalyzer, "panics")
}

func TestLogFatalAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), LogFatalAnalyzer, "logfatal", "logfatallib")
}

func TestOsExitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), OsExitAnalyzer, "osexit", "osexitlib")
}

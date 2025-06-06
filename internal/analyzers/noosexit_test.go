package analyzers

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoOSExitMain(t *testing.T) {
	tests := []struct {
		name   string
		folder string
	}{
		{
			name:   "valid",
			folder: "valid",
		},
		{
			name:   "Main func have exit",
			folder: "mainfuncexit",
		},
		{
			name:   "Not main func has os.exit",
			folder: "notmainfuncexit",
		},
		{
			name:   "Not main package",
			folder: "notmainpackage",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testdata := analysistest.TestData()
			analysistest.Run(t, testdata, NoOSExitMainAnalyzer, tt.folder)
		})
	}
}

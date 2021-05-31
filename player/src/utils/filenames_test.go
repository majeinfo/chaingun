package utils

import (
	"os"
	"testing"
)

type test_filename struct {
	filename string
	playbookdir string
	result string
}

type test_output_filename struct {
	output_dir string
	output_type string
	result string
	dir string
}

func TestComputeFilename(t *testing.T) {
	tests := []test_filename{
		{"test.txt", "", "/test.txt"},
		{"/tmp/test.txt", "", "/tmp/test.txt"},
		{"test.txt", "/tmp", "/tmp/test.txt"},
	}

	for _, test := range tests {
		if result := ComputeFilename(test.filename, test.playbookdir); result != test.result {
			t.Errorf("ComputeFilename(%s, %s) must return '%s', but returned '%s'", test.filename, test.playbookdir, test.result, result)
		}
	}
}

func TestComputeOutputFilename(t *testing.T) {
	d, _ := os.Getwd()
	tests := []test_output_filename{
		{"", "", d + "/results/" + "data.csv", d + "/results/"},
		{"/tmp", "", "/tmp/data.csv", "/tmp/"},
		{"/tmp/", "", "/tmp/data.csv", "/tmp/"},
		{"output", "", "output/data.csv", "output/"},
	}

	for _, test := range tests {
		if outputfile, dir := ComputeOutputFilename(test.output_dir, test.output_type); outputfile != test.result || dir != test.dir {
			t.Errorf("ComputeOutputFilename(%s, %s) must return '%s', '%s', but returned '%s', '%s'", test.output_dir, test.output_type, test.result, test.dir, outputfile, dir)
		}
	}
}

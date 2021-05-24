package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// Compute the real filename when given as a relative name :
// in this case, make sure the relative name is relative to the playbook directory !
func ComputeFilename(filename string, playbookDir string) string {
	log.Debugf("ComputeFilename for %s", filename)
	// Check if filename is absolute or not
	if filename[0] != '/' {
		filename = playbookDir + "/" + filename
	}

	log.Debugf("Computed filename is %s", filename)
	return filename
}

// Compute the name of the output file (/path/to/data.csv)
func ComputeOutputFilename(output_dir string, output_type string) (string, string) {
	var outputfile string
	var dir string

	if output_dir == "" {
		d, _ := os.Getwd()
		dir = d + "/results"
	} else {
		dir = output_dir
	}
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	//outputfile = dir + "data." + output_type
	// Always generate a CSV file
	outputfile = dir + "data.csv"

	return outputfile, dir
}

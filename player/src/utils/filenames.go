package utils

import (
	log "github.com/sirupsen/logrus"
)

// Compute the real filename when given as a relative name :
// in this case, make sure the relative name is relative to the plaubook directory !
func ComputeFilename(filename string, playbookDir string) string {
	log.Debugf("ComputeFilename for %s", filename)
	// Check if filename is absolute or not
	if filename[0] != '/' {
		filename = playbookDir + "/" + filename
	}

	log.Debugf("Computed filename is %s", filename)
	return filename
}

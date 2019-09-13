package action

import (
	"gopkg.in/yaml.v2"
	"path"

	"github.com/majeinfo/chaingun/config"
	log "github.com/sirupsen/logrus"
)

// Create a Playbook from the YAML data
func CreatePlaybook(scriptFile *string, data []byte, playbook *config.TestDef, actions *[]FullAction) bool {
	err := yaml.UnmarshalStrict([]byte(data), playbook)
	if err != nil {
		log.Fatalf("YAML error: %v", err)
	}
	log.Debug("Playbook:")
	log.Debug(playbook)

	if !config.ValidateTestDefinition(playbook) {
		return false
	}

	// Add the Feeder filename in the list
	embedded_files = make([]string, 0)
	if playbook.DataFeeder.Type != "" {
		embedded_files = append(embedded_files, playbook.DataFeeder.Filename)
	}

	var isValid bool
	*actions, isValid = BuildActionList(playbook, path.Dir(*scriptFile))
	if !isValid {
		return false
	}
	//log.Debug("Tests Definition:")
	//log.Debug(playbook)

	return true
}

// Add a filename to the list of embedded filnames
func addEmbeddedFilename(fname string) {
	embedded_files = append(embedded_files, fname)
}

// Return the list of embedded filenames
func GetEmbeddedFilenames() []string {
	return embedded_files
}

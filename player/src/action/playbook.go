package action

import (
	"gopkg.in/yaml.v2"
	"path"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

// Create a Playbook from the YAML data
func CreatePlaybook(scriptFile string, data []byte, playbook *config.TestDef, pre_actions *[]FullAction, actions *[]FullAction, post_actions *[]FullAction) bool {
	gpScriptDir = path.Dir(scriptFile)
	log.Debugf("ScriptDir=%s", gpScriptDir)

	err := yaml.UnmarshalStrict([]byte(data), playbook)
	if err != nil {
		//log.Fatalf("YAML error: %v", err)
		log.Errorf("YAML error: %v", err)
		return false
	}
	log.Debug("Playbook:")
	log.Debug(playbook)

	if !config.ValidateTestDefinition(playbook) {
		return false
	}

	// Add the Feeder filename in the list
	embedded_files = make([]string, 0)
	if playbook.DataFeeder.Type != "" {
		addEmbeddedFilename(playbook.DataFeeder.Filename)
	}

	var isValid bool
	*pre_actions, *actions, *post_actions, isValid = BuildActionList(playbook)
	if !isValid {
		return false
	}
	//log.Debug("Tests Definition:")
	//log.Debug(playbook)

	return true
}

// Add a filename to the list of embedded filnames
func addEmbeddedFilename(fname string) {
	embedded_files = append(embedded_files, utils.ComputeFilename(fname, gpScriptDir))
}

// Return the list of embedded filenames
func GetEmbeddedFilenames() []string {
	return embedded_files
}

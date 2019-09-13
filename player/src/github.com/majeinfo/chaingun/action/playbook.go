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

	var isValid bool
	*actions, isValid = BuildActionList(playbook, path.Dir(*scriptFile))
	if !isValid {
		return false
	}
	log.Debug("Tests Definition:")
	log.Debug(playbook)

	return true
}

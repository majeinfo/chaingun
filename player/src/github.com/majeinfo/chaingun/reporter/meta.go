package reporter

import (
	"gopkg.in/yaml.v2"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	METAFILE_NAME = "meta.yml"
)

type Metadata struct {
	StartTime  time.Time `yaml:"starttime"`
	EndTime    time.Time `yaml:"endtime"`
	ScriptName []string  `yaml:"scriptname"`
}

// Write the file with the tests metadata
func WriteMetadata(starttime time.Time, endtime time.Time, dirname string, scriptnames []string) error {
	fileName := dirname + "/" + METAFILE_NAME
	f, err = os.Create(fileName)
	if err != nil {
		log.Fatalf("Cannot create output file %s: %s", fileName, err)
	}
	defer f.Close()

	t := Metadata{
		StartTime:  starttime,
		EndTime:    endtime,
		ScriptName: scriptnames,
	}

	d, err := yaml.Marshal(&t)
	if err != nil {
		log.Errorf("WriteMetadata failed to Marshal data: %v", err)
		return err
	}

	f.Write(d)

	return nil
}

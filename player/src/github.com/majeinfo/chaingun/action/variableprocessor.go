package action

import (
	"regexp"
	"strings"
	"net/url"

    log "github.com/sirupsen/logrus"	
)

var re = regexp.MustCompile("\\$\\{([a-zA-Z0-9]{0,})\\}")

func SubstParams(sessionMap map[string]string, textData string) string {
	if strings.ContainsAny(textData, "${") {
		res := re.FindAllStringSubmatch(textData, -1)
		for _, v := range res {
			log.Debugf("sessionMap[%s]=%s", v[1], sessionMap[v[1]])
			if _, err := sessionMap[v[1]]; !err {
				log.Errorf("Variable ${%s} not set", v[1])
			}
			textData = strings.Replace(textData, "${" + v[1] + "}", url.QueryEscape(sessionMap[v[1]]), 1)
		}
		return textData
	} else {
		return textData
	}
	return textData
}

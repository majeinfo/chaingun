package action

import (
	"regexp"
	"strings"
	"net/url"

    log "github.com/sirupsen/logrus"	
)

var re = regexp.MustCompile("\\$\\{([a-zA-Z0-9_]{0,})\\}")

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
	} 

	return textData
}

var escape_re = regexp.MustCompile("\\$%7B([a-zA-Z0-9]{0,})%7D")

func RedecodeEscapedPath(escaped_url string) string {
	unescaped_url := escaped_url

	if strings.ContainsAny(escaped_url, "$%7B") {
		res := escape_re.FindAllStringSubmatch(unescaped_url, -1)
		for _, v := range res {
			log.Debugf(v[1])
			unescaped_url = strings.Replace(unescaped_url, "$%7B" + v[1] + "%7D", "${" + v[1] + "}", 1)
		}
		return unescaped_url
	}

	return unescaped_url
}

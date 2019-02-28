package action

import (
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

var re = regexp.MustCompile("\\$\\{([a-zA-Z0-9_]{0,})\\}")

// SubstParams compute the result of variables interpolation
func SubstParams(sessionMap map[string]string, textData string) string {
	if strings.ContainsAny(textData, "${") {
		res := re.FindAllStringSubmatch(textData, -1)
		for _, v := range res {
			log.Debugf("sessionMap[%s]=%s", v[1], sessionMap[v[1]])
			if _, err := sessionMap[v[1]]; !err {
				log.Errorf("Variable ${%s} not set", v[1])
			}
			textData = strings.Replace(textData, "${"+v[1]+"}", url.QueryEscape(sessionMap[v[1]]), 1)
		}
		return textData
	}

	return textData
}

var escapeRe = regexp.MustCompile("\\$%7B([a-zA-Z0-9]{0,})%7D")

// RedecodeEscapedPath gives the canonical value of un escaped path
func RedecodeEscapedPath(escapedURL string) string {
	unescapedURL := escapedURL

	if strings.ContainsAny(escapedURL, "$%7B") {
		res := escapeRe.FindAllStringSubmatch(unescapedURL, -1)
		for _, v := range res {
			log.Debugf(v[1])
			unescapedURL = strings.Replace(unescapedURL, "$%7B"+v[1]+"%7D", "${"+v[1]+"}", 1)
		}
		return unescapedURL
	}

	return unescapedURL
}

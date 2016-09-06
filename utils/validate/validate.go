package validate

import (
	"regexp"
)

var nameregexp = `^[A-Za-z0-9][\w]{0,254}$`
var reporegexp = `^[A-Za-z0-9][\w-]{0,254}$`
var appregexp = `^[A-Za-z0-9][\w.-]{0,254}$`
var tagregexp = `^[\w][\w.-]{0,127}$`
var digestregexp = `^[A-Za-z][A-Za-z0-9]*(?:[-_+.][A-Za-z][A-Za-z0-9]*)*[:][[:xdigit:]]{32,}$`

func IsNameValid(name string) bool {
	return regexp.MustCompile(nameregexp).MatchString(name)
}

func IsRepoValid(repo string) bool {
	return regexp.MustCompile(reporegexp).MatchString(repo)
}

func IsAppValid(app string) bool {
	return regexp.MustCompile(appregexp).MatchString(app)
}

func IsTagValid(tag string) bool {
	return regexp.MustCompile(tagregexp).MatchString(tag)
}

func IsDigestValid(digest string) bool {
	return regexp.MustCompile(digestregexp).MatchString(digest)
}

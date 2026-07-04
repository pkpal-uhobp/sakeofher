package service

import "regexp"

var nonRemnaUsernameChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

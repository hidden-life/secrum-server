package messages

import "strings"

func intersectAllowed(systemAllowed func(string) bool, groupAllowed, userAllowed []string, mimeType string) bool {
	// system
	if !systemAllowed(mimeType) {
		return false
	}

	// group
	if len(groupAllowed) > 0 && !containsMimeType(groupAllowed, mimeType) {
		return false
	}

	// user (if set)
	if len(userAllowed) > 0 && !containsMimeType(userAllowed, mimeType) {
		return false
	}

	return true
}

func containsMimeType(list []string, mimeType string) bool {
	for _, v := range list {
		if strings.EqualFold(v, mimeType) {
			return true
		}
	}
	
	return false
}

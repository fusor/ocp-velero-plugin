package common

import (
	"fmt"
	"strings"
)

// ReplaceImageRefPrefix replaces an image reference prefix with newPrefix.
// If the input image reference does not start with oldPrefix, an error is returned
func ReplaceImageRefPrefix (s, oldPrefix, newPrefix string) (string, error) {
	refSplit := strings.SplitN(s, "/", 2)
	if len(refSplit) != 2 {
		err := fmt.Errorf("image reference [%v] does not have prefix [%v]", s, oldPrefix)
		return "", err
	}
	if refSplit[0] != oldPrefix {
		err := fmt.Errorf("image reference [%v] does not have prefix [%v]", s, oldPrefix)
		return "", err
	}
	return fmt.Sprintf("%s/%s", newPrefix, refSplit[1]), nil
}

// HasImageRefPrefix returns true if the input image reference begins with
// the input prefix followed by "/"
func HasImageRefPrefix(s, prefix string) bool {
	refSplit := strings.SplitN(s, "/", 2)
	if len(refSplit) != 2 {
		return false
	}
	return refSplit[0] == prefix
}

package configmapper

import (
	"fmt"
	"regexp"
	"strconv"
)

const RegexArr = `^(.+)\[(\d+)\]$`

// CheckNameIsArrayAndGetIndex
// this function checks to see if a key ends in [0-9] number or not
// Used for building up array of configs directly in the config file
// Eliminating the need to post-boot parsing of the array configs
func CheckNameIsArrayAndGetIndex(key string) (originalName string, is bool, index int) {
	if matched, err := regexp.Compile(RegexArr); err == nil {
		res := matched.FindSubmatch([]byte(key))
		if res != nil && len(res) > 2 {
			var num, err = strconv.ParseInt(string(res[2]), 10, 64)
			if err != nil || num < 0 {
				fmt.Printf("failed to parse array key name, the value [%s] must be an uint value", string(res[1]))
				return "", false, -1
			}
			return string(res[1]), true, int(num)
		}
	}
	return "", false, -1
}

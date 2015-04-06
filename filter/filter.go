package filter

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func GetFilterList(omitString string, verbose bool) []string {
	// POST: FCTVAL== slice of strings (paths and files) to filter
	filterList := []string{} // filtered list to be returned

	// Add .cfignore files to filterList
	content, err := ioutil.ReadFile(".cfignore")

	if err != nil && verbose {
		fmt.Println("[ Info: ", err, "]")
	} else {
		lines := strings.Split(string(content), "\n") // get each line in .cfignore

		if verbose && len(lines) > 0 {
			fmt.Println("[ Info: using .cfignore ] \nContents: ")
			for _, val := range lines {
				fmt.Println(val)
			}
			fmt.Println("")
		} else if len(lines) > 0 {
			fmt.Println("[ Info: using .cfignore ]")
		}

		filterList = append(filterList, lines[0:]...)
	}

	// Add the path from the --omit param to filterList
	allOmits := strings.Split(omitString, ";")

	// Add omitted strings to the filter list
	filterList = append(filterList, allOmits[0:]...)

	var returnList []string // filtered strings to be returned

	// Remove any trailing forward slashes in the filterList[ex: app/ becomes app]
	for i, _ := range filterList {
		filterList[i] = strings.TrimSpace(filterList[i])

		if filterList[i] != "" {
			filterList[i] = strings.TrimPrefix(filterList[i], "/")
			filterList[i] = "/" + filterList[i]
			filterList[i] = strings.TrimSuffix(filterList[i], "/")

			returnList = append(returnList, filterList[i])
		}
	}

	return returnList
}

func CheckToFilter(appPath, rootWorkingDirectory string, filterList []string) bool {
	appPath = strings.TrimSuffix(appPath, "/")
	comparePath1 := strings.TrimPrefix(appPath, rootWorkingDirectory)
	/*fmt.Println("\nCompareTo:	   ", comparePath1)*/
	for _, item := range filterList {

		// ignore files in ignore list and the cfignore file
		if comparePath1 == item {
			return true
		}
	}

	return false
}

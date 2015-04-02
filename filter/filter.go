package filter

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func GetFilterList(omitString string, verbose bool) []string {
	// POST: FCTVAL== slice of strings (paths and files) to filter
	var filterList []string // filtered list to be returned

	// Add .cfignore files to filterList
	content, err := ioutil.ReadFile(".cfignore")
	if err != nil && verbose {
		fmt.Println("[ Info: ", err, "]")
	} else {
		lines := strings.Split(string(content), "\n") // get each line in .cfignore

		// Remove any leading forward slashes
		for i := 0; i < len(lines); i++ {
			lines[i] = strings.TrimPrefix(lines[i], "/")
		}

		filterList = append(filterList, lines[0:]...)

		// remove empty strings that we got from the last line
		if len(filterList) > 0 && filterList[len(filterList)-1] == "" {
			filterList = filterList[:len(filterList)-1]
		}
	}

	// Add the path from the --omit param to filterList
	if omitString != "" {

		allOmits := strings.Split(omitString, ";")

		// Parse for each path and remove leading forward slashes
		for i := 0; i < len(allOmits); i++ {
			allOmits[i] = strings.TrimSpace(allOmits[i])
			allOmits[i] = strings.TrimPrefix(allOmits[i], "/")
		}
		filterList = append(filterList, allOmits[0:]...)
	}

	var returnList []string // filtered strings to be returned

	// Remove any trailing forward slashes in the filterList[ex: app/ becomes app]
	for i, _ := range filterList {
		filterList[i] = strings.TrimSuffix(filterList[i], "/")
		filterList[i] = "/" + filterList[i]

		// don't include any empty strings, which only have a forward slash
		if strings.TrimSpace(filterList[i]) != "/" { //there are multiple checks for empty string above, seems redundant, maybe append all and check whole list here?
			returnList = append(returnList, filterList[i]) //just append anything >1 character after trimming space
		}
	}

	return returnList
}

func CheckToFilter(appPath, rootWorkingDirectory string, filterList []string) bool {
	appPath = strings.TrimSuffix(appPath, "/")
	comparePath1 := strings.TrimPrefix(appPath, rootWorkingDirectory)
	/*fmt.Println("\nCompareTo:	   ", comparePath1)*/
	for _, item := range filterList {

		/*fmt.Println("	filterItem:", item)
		fmt.Println("	equal:", comparePath1 == item)*/
		if comparePath1 == item {
			return true
		}
	}

	return false
}

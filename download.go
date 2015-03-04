package main

import (
	"fmt"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/mgutz/ansi"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
    "flag"
    "strconv"
)

/*
*	This is the struct implementing the interface defined by the core CLI. It can
*	be found at  "github.com/cloudfoundry/cli/plugin/plugin.go"
 */
type downloadPlugin struct{}

// error struct that allows appending error messages
type cliError struct {
	err    error
	errMsg string
}

var (
	rootWorkingDirectory string
	appName              string
	instance             string
	verbose              bool
	failedDownloads      []string
	filesDownloaded      int
)

// global wait group for all download threads
var wg sync.WaitGroup



func getFilterList(omitString string) []string {
    var filterList []string                            // filtered list to be returned 
    
    // Add .cfignore files to filterList
    content, err := ioutil.ReadFile(".cfignore")
    if err != nil {
        fmt.Println(err)
    } else {
        lines := strings.Split(string(content), "\n")
        filterList = append(filterList,lines[0:]...)
        if len(filterList) > 0 && filterList[len(filterList)-1] == "" { 
            filterList = filterList[:len(filterList)-1]
        }
    }
    
    if omitString != "" {
        filterList = append(filterList, omitString)                  // add -omit param to filterList
    }
    
    // Remove any trailing forward slashes in the filterList[ex: app/ becomes app]
    for i, _ := range filterList {
        filterList[i] = strings.TrimSuffix(filterList[i], "/")
        filterList[i] = "/"+filterList[i]
    }
    
    return filterList
}
/*
*	This function must be implemented by any plugin because it is part of the
*	plugin interface defined by the core CLI.
*
*	Run(....) is the entry point when the core CLI is invoking a command defined
*	by a plugin. The first parameter, plugin.CliConnection, is a struct that can
*	be used to invoke cli commands. The second paramter, args, is a slice of
*	strings. args[0] will be the name of the command, and will be followed by
*	any additional arguments a cli user typed in.
*
*	Any error handling should be handled with the plugin itself (this means printing
*	user facing errors). The CLI will exit 0 if the plugin exits 0 and will exit
*	1 should the plugin exits nonzero.
 */
func (c *downloadPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	// start time for download timer
	start := time.Now()
    proceed := true
	
    if len(args) < 2 {
			fmt.Println("\nError: Missing App Name")
			proceed = false
    }
    
    
    
    // Create flag
    omitp := flag.String("omit", "", "--omit path/to/some/file")
    overWritep := flag.Bool("overwrite", false, "--overwrite")
    maxRoutinesp := flag.Int("routines", 200, "--routines [numOfRoutines]")
    instancep := flag.Int("i", 0, "-i [instanceNum]")
    verbosep := flag.Bool("verbose", false, "--verbose")
    
    
    copyOfArgs := make([]string, len(args))         // need to copy args[] for later as they will be overwritten
    
    for i := 0; i < len(args); i++ {
           copyOfArgs[i] = args[i]
    }
    
    os.Args = append(os.Args[:1], args[2:]...)  // flag package parses os.Args
    appName = copyOfArgs[1]
    
    
    
    // check for misplaced flags
    if strings.HasPrefix(appName, "-") || strings.HasPrefix(appName, "--") {
        fmt.Println("\nError: App name begins with '-' or '--'. correct flag usage: 'cf download APP_NAME [--flags]'")
        proceed = false
    }
    
    
    // make sure the flags have valid input
    for i := 2; i < len(copyOfArgs); i++ {
        if len(args) > 2 && !strings.HasPrefix(copyOfArgs[2],"-") {    // has specified app dir/file

        } else {                                                       // no specified app dir/file

            temp := strings.TrimPrefix(copyOfArgs[i],"-")
            temp = strings.TrimPrefix(temp,"-")
            
            switch temp {
			case "omit": 
			//	fmt.Printf("omit not recognized")
			case "verbose":
			//	fmt.Printf("verbose not recognized")
			case "overwrite":
			//	fmt.Printf("overwrite not recognized")
			case "i":
			//	fmt.Printf("i not recognized")
            case "routines":
          //      fmt.Printf("routines not recognized")
            default:
             //   fmt.Printf("Argument not recognized")
			}
        }
        
    }
    
    // Parse flags
    flag.Parse()
    
	// flag variables
	maxRoutines := *maxRoutinesp
	overWrite := *overWritep
	instance = strconv.Itoa(*instancep)
	verbose = *verbosep

	runtime.GOMAXPROCS(maxRoutines)

	
    if proceed == false {
        os.Exit(1)   
    } else {
        filterList := getFilterList(*omitp)             // get list of things to not download
		workingDir, err := os.Getwd()
		check(cliError{err: err, errMsg: "Called by: Run"})
		rootWorkingDirectory = workingDir + "/" + appName + "-download/"



		// ensure cf_trace is disabled, otherwise parsing breaks
		if os.Getenv("CF_TRACE") == "true" {
			fmt.Println("\nError: environment variable CF_TRACE is set to true. This prevents download from succeeding.")
			return
		}

		// prevent overwriting files
		if exists(rootWorkingDirectory) && overWrite == false {
			fmt.Println("\nError: destination path", rootWorkingDirectory, "already exists and is not an empty directory. Delete it or use 'cf download APP_NAME --overwrite'")
			return
		}

		// append path if provided as arguement
		startingPath := "/"
		if len(args) > 2 && !strings.HasPrefix(copyOfArgs[2],"-"){
			startingPath = copyOfArgs[2]
			if !strings.HasSuffix(startingPath, "/") {
				startingPath += "/"
			}
            rootWorkingDirectory += startingPath
		}

		// parse the directory
		files, dirs := execParseDir(startingPath)

		fmt.Printf("Files completed: %d", filesDownloaded)

		// stop consoleWriter
		quit := make(chan int)
		// disable consoleWriter if verbose
		if verbose == false {
			go consoleWriter(quit)
		}

		// Start the download
		wg.Add(1)
		download(files, dirs, startingPath, rootWorkingDirectory, filterList)

		// Wait for download goRoutines
		wg.Wait()

		// stop console writer
		if verbose == false {
			quit <- 0
		}

		// let user know if any files were inaccessible
		if len(failedDownloads) == 1 {
			fmt.Println("")
			fmt.Println(len(failedDownloads), "file was not downloaded (inaccessible or corrupt):")
			printSlice(failedDownloads)
		} else if len(failedDownloads) > 1 {
			fmt.Println("")
			fmt.Println(len(failedDownloads), "files were not downloaded (inaccessible or corrupt):")
			printSlice(failedDownloads)
		}

		// display runtime
		elapsed := time.Since(start)
		fmt.Printf("\nDownload time: %s\n", elapsed)

		msg := ansi.Color(appName+" Successfully Downloaded!", "green+b")
		fmt.Println(msg)
	}
}

/*
*	consoleWriter prints the current number of files downloaded. It is polled every 350 milleseconds
* 	disabled if using verbose flag.
 */
func consoleWriter(quit chan int) {
	count := 0
	for {
		select {
		case <-quit:
			return
		default:
			switch count = (count + 1) % 4; count {
			case 0:
				fmt.Printf("\rFiles completed: %d \\ ", filesDownloaded)
			case 1:
				fmt.Printf("\rFiles completed: %d | ", filesDownloaded)
			case 2:
				fmt.Printf("\rFiles completed: %d / ", filesDownloaded)
			case 3:
				fmt.Printf("\rFiles completed: %d --", filesDownloaded)
			}
			time.Sleep(350 * time.Millisecond)
		}
	}
}

/*
*	execParseDir() uses os/exec to shell out commands to cf files with the given readPath. The returned
*	text contains file and directory structure which is then parsed into two slices, dirs and files. dirs
*	contains the names of directories in readPath, files contians the file names. dirs and files are returned
* 	to be downloaded by download() and downloadFile() respectively.
 */
func execParseDir(readPath string) ([]string, []string) {
	// make the cf files call using exec
	cmd := exec.Command("cf", "files", appName, readPath, "-i", instance)
	output, err := cmd.CombinedOutput()
	dirSlice := strings.SplitAfterN(string(output), "\n", 3)

	// check for invalid or missing app
	if strings.Contains(dirSlice[1], "not found") {
		errmsg := ansi.Color("Error: "+appName+" app not found (check space and org)", "red+b")
		fmt.Println(errmsg)
	}

	// this usually gets called when an app is not running and you attempt to download it.
	dir := dirSlice[2]
	if strings.Contains(dir, "error code: 190001") {
		errmsg := ansi.Color("App not found, possibly not yet running", "red+b")
		fmt.Println(errmsg)
		check(cliError{err: err, errMsg: "App not found"})
	}

	// handle an empty directory
	if strings.Contains(dir, "No files found") {
		return nil, nil
	} else {
		check(cliError{err: err, errMsg: "Called by: ExecParseDir [cf files " + appName + " " + readPath + "]"})
	}

	// parse the returned output into files and dirs slices
	filesSlice := strings.Fields(dir)
	var files, dirs []string
	for i := 0; i < len(filesSlice); i += 2 {
		if strings.HasSuffix(filesSlice[i], "/") {
			dirs = append(dirs, filesSlice[i])
		} else {
			files = append(files, filesSlice[i])
		}

	}
	return files, dirs
}

/*
*	downloadFile() takes a 'readPath' which corresponds to a file in the cf app. The file is
*	downloaded using the os/exec library to call cf files with the given readPath. The output is
*	written to a file at writePath.
 */
func downloadFile(readPath, writePath string, fileDownloadGroup *sync.WaitGroup) error {
	defer fileDownloadGroup.Done()

	// call cf files using os/exec
	cmd := exec.Command("cf", "files", appName, readPath, "-i", instance)
	output, err := cmd.CombinedOutput()
	file := strings.SplitAfterN(string(output), "\n", 3)

	// check for invalid file error.
	// some files are inaccesible from the cf files (reasons unknown) this is rare but we need to
	// alert users if it occurs
	fileAsString := file[2]
	if strings.Contains(file[1], "FAILED") {
		errormsg := ansi.Color(" Server Error: '"+readPath+"' not downloaded", "red")
		failedDownloads = append(failedDownloads, errormsg)
		fmt.Println(errormsg)
		return nil
	} else {
		// check for other errors
		check(cliError{err: err, errMsg: "Called by: downloadFile 1"})
	}

	// write downloaded file to writePath
	err = ioutil.WriteFile(writePath, []byte(fileAsString), 0644)
	check(cliError{err: err, errMsg: "Called by: downloadFile 2"})
	if verbose {
		fmt.Printf("Wrote file: %s\n", readPath)
	} else {
		// increment download counter for commandline display
		// see consoleWriter()
		filesDownloaded++
	}

	return nil
}

func checkToFilter(appPath string, filterList []string) bool {
    appPath = strings.TrimSuffix(appPath, "/")
    comparePath1 := strings.TrimPrefix(appPath, rootWorkingDirectory)
    
    for _, item := range filterList {
       if comparePath1 == item {
           return true
       }
    }
    
    return false              
}

/*
*	given file and directory names, download() will download the files from
* 	'readPath' and write them to disk on the 'writepath'.
* 	the function calls it's self recursively for each directory as it travels down the tree.
* 	Each call runs on a seperate go routine and and calls a go routine for every
* 	file download.
 */
func download(files, dirs []string, readPath, writePath string, filterList []string) error {
	defer wg.Done()

	//create dir if does not exist
	err := os.MkdirAll(writePath, 0755)
	check(cliError{err: err, errMsg: "Called by: download"})

	// download each file
	for _, val := range files {
		fileWPath := writePath + val
		fileRPath := readPath + val
        
        if checkToFilter(fileRPath,filterList) {
			continue
		}
        
		wg.Add(1)
		go downloadFile(fileRPath, fileWPath, &wg)
	}

	// call download on every sub directory
	for _, val := range dirs {
		dirWPath := writePath + val
		dirRPath := readPath + val
        
        if checkToFilter(dirRPath,filterList) {
			continue
		}
        
		err := os.MkdirAll(dirWPath, 0755)
		check(cliError{err: err, errMsg: "Called by: download"})

		files, dirs = execParseDir(dirRPath)

		wg.Add(1)
		go download(files, dirs, dirRPath, dirWPath, filterList)
	}
	return nil
}

/*
*	This function must be implemented as part of the	plugin interface
*	defined by the core CLI.
*
*	GetMetadata() returns a PluginMetadata struct. The first field, Name,
*	determines the name of the plugin which should generally be without spaces.
*	If there are spaces in the name a user will need to properly quote the name
*	during uninstall otherwise the name will be treated as seperate arguments.
*	The second value is a slice of Command structs. Our slice only contains one
*	Command Struct, but could contain any number of them. The first field Name
*	defines the command `cf basic-plugin-command` once installed into the CLI. The
*	second field, HelpText, is used by the core CLI to display help information
*	to the user in the core commands `cf help`, `cf`, or `cf -h`.
 */
func (c *downloadPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "download",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 1,
		},
		Commands: []plugin.Command{
			plugin.Command{
				Name:     "download",
				HelpText: "Download contents of targeted directory",

				// UsageDetails is optional
				// It is used to show help of usage of each command
				UsageDetails: plugin.Usage{
					Usage: "cf download APP_NAME [PATH] [--flags]",
					Options: map[string]string{
						"overwrite": "overwrite files",
						"verbose":   "verbose",
						"omit":      "omit directory or file",
						"routines":  "max number of concurrent subroutines (default 200)",
						"i":         "instance",
					},
				},
			},
		},
	}
}

// error check function
func check(e cliError) {
	if e.err != nil {
		fmt.Println("\nError: ", e.err)
		if e.errMsg != "" {
			fmt.Println("Message: ", e.errMsg)
		}
		os.Exit(1)
	}
}

// prints slices in readable format
func printSlice(slice []string) error {
	for index, val := range slice {
		fmt.Println(index+1, ": ", val)
	}
	return nil
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	check(cliError{err: err, errMsg: "Called by: exists"})
	return false
}

/*
* Unlike most Go programs, the `Main()` function will not be used to run all of the
* commands provided in your plugin. Main will be used to initialize the plugin
* process, as well as any dependencies you might require for your
* plugin.
 */
func main() {
	// Any initialization for your plugin can be handled here

	// Note: The plugin's main() method is invoked at install time to collect
	// metadata. The plugin will exit 0 and the Run([]string) method will not be
	// invoked.
	plugin.Start(new(downloadPlugin))
	// Plugin code should be written in the Run([]string) method,
	// ensuring the plugin environment is bootstrapped.
}

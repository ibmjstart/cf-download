# CF DOWNLOAD
### A Cloud Foundry cli plugin for downloading your application contents after staging



##Installation
#####Install from CLI (Recommended)
  ```
  $ cf add-plugin-repo CF-Community http://plugins.cloudfoundry.org/
  $ cf install-plugin targets -r CF-Community
  ```
  
##### Install from binary
1. download binary (See Download Section below)
2. **cd path/to/downloaded/binary**
3. If you've already installed the plugin and are updating, you must first run **cf uninstall-plugin download**
4. Then install the plugin with **cf install-plugin cf-download** 
	* If you get a permission error run: **chmod +x cf-download** on the binary
5. Verify the plugin installed by looking for it with **cf plugins** 

##### Download Binaries

###### Mac:     [64-bit](https://github.com/ibmjstart/cf-download/blob/master/binaries/darwin/amd64/cf-download?raw=true)   
###### Windows: [64-bit](https://github.com/ibmjstart/cf-download/blob/master/binaries/windows/amd64/cf-download.exe?raw=true)    
###### Linux:   [64-bit](https://github.com/ibmjstart/cf-download/blob/master/binaries/linux/amd64/cf-download?raw=true)

***

## Usage

cf download APP_NAME [PATH] [--overwrite] [--verbose] [--omit omitted_path] [-i instance]

The downloaded app files will put put in a new directory "APP_NAME-download" that's created within your working directory.

### Path Argument
The path argument is optional but if included, should come immediately after the app name. It determines the starting directory that all the files will be downloaded from. By default, the entire app is downloaded starting from the root. However if desired, one could use **some/starting/path** to only download files within the **some/starting/path** directory. Additionally, the path can point to a single file to be downloaded. Note: this works similarly to "cf files [path]".

### Flags:
1. The **--overwrite** flag is needed if the download directory, "APP_NAME-download", is already taken. Using the flag, that directory will be overwritten.
2. The **--verbose** flag is used to see more detailed output as the downloads are happening.
3. The **--omit [omitted_path]** flag is useful when certain files or directories are not wanted. You can exclude a file by typing **--omit path/to/file**. Multiple things can be omitted by delimiting the paths with semicolons and putting quotes around the entire parameter like so: **--omit "path/to/file; another/path/to/file"**
5. The **-i [instance]** flag will download from the given app instance. By default, the instance number is 0.

***

## Improving performance: 
Projects usually have a enormous amount of dependencies installed by package managers, we highly recommend not downloading these dependencies. Using the --omit flag, you can avoid downloading these dependencies and significantly reduce download times.

#### Java/Liberty:
We highly reccomend you not download the app/.java and app/.liberty directories in your java/liberty projects. They are very large and contain many permission issues the prevent proper downloads. It is best to omit them. 

#### Node.js:
npm will download dependencies to the node_modules folder in the app directory. By omitting app/node_modules you will greatly decrease download times. You can run npm install locally on your package.json after completing a download. 

#### PHP:
Composer is a popular PHP package manager that installs dependencies to a folder called vendors. It is reccomended you omit this folder from the download to ensure a quick and error free download. example: **--omit <path_to_vendors>/vendors** 

***

## Notes and FAQ:  
#### .cfignore:
All directories and files within the .cfignore file will be omitted. Each entry should be on its own line and that the .cfignore file must be in the same working directory. Instead of using many --omit parameters, it's easier to use the .cfignore file.

#### Stuck Download:  
In case your download seems to be stuck, we recommend terminating and redownloading using the --verbose flag. When the download stalls you can see which files were being downloaded and what could be causing the issue. 

#### Downloading Jar files:
Projects containing jar files can trigger antivirus software while being downloaded. you can either temporarily disable network antivirus protection or exclude directories containing jar files.

#### I am getting a lot of 502 errors, why?:
On rare occasions the cf cli api that the plugin uses can get overburdened by the plugin. This will display a lot of 502 error messages to the command line. The best thing to do in this case is wait a couple minutes and try again later. The Api will hopefully return to full capacity and allow downloads to complete. In the unlikely case that you experience this often, create an issue on this repo and we can explore solutions.  

#### Error: "App not found, or the app is in stopped state (This can also be caused by api failure)":
This error is caused when the cf cli api fails. Best solution is to wait and try again, when the api recovers.

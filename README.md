# CF DOWNLOAD
### A Cloud Foundry cli plugin for downloading  

## Install 
1. download binary (See Download Section)
2. **cd path/to/downloaded/binary**
3. Only if updating: run **cf uninstall-plugin cf-download** first
3. **cf install-plugin cf-download** note: may require sudo because of download file permissions
4. make sure plugin shows up in **cf plugins** 

## Download

#### Mac:       [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/darwin/386/cf-download) | [amd 64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/darwin/amd64/cf-download)   
#### Windows:   [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/blob/master/binaries/windows/386/cf-download.exe) | [amd64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/blob/master/binaries/windows/amd64/cf-download.exe)    
#### Linux:     [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/386/cf-download) | [amd64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/amd64/cf-download) | [arm](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/arm/cf-download)

## Usage

cf download APP_NAME [PATH] [--overwrite] [--verbose] [--omit ommited_path] [--routines max_routines] [-i instance]

path and flags are optional, setting a path will start the download from the specified path instead of the app root.
Use cf help download to see options

***

## Improving performance: 
Projects usually have a enormous amount of dependencies installed by package managers, we highly reccomend not downloading these dependencies. using the --omit flag you can avoid downloading these dependencies and significantly reduce download times.

#### Java/Liberty:
We highly reccomend you not download the app/.java and app/.liberty directories in your java/liberty projects. They are very large and contain many permission issues the prevent proper downloads. It is best to omit them using an omit flag to avoid them. Syntax for multiple omits (comma sseperated paths starting from project root) **--omit "app/.java, app/.liberty"** 

#### Node.js:
npm will download dependencies to the node_modules folder in the app directory. By omitting app/node_modules you will greatly decrease download times. You can run npm install locally on your package.json after completing a download. This is a quick and easy way to get these files. Syntax for omit: **--omit app/node_modules**

#### PHP:
Composer is a popular PHP package manager that installs dependencies to a folder called vendors. It is reccomended you omit this folder from the download to ensure a quick and error free download. example: **--omit <path_to_vendors>/vendors** Composer can be run locally to download these dependencies later.

***

## Notes and FAQ:  
#### Limiting Resources:  
The download plugin is extremely concurrent and can use up a lot of cpu power. You can reduce the max number of concurrent routines with the --routines flag. Reducing the number of routines reduces the cpu load at the cost of speed.

#### .cfignore:
you can create a .cfignore file in your current directory to tell the plugin which paths should be ignored dunring the download. This helps avoid having to use a bunch of --omit flags

#### Stuck Download:  
In case the download seems to be stuck, we recommend terminating and redownloading using the verbose flag. When the download stalls you can see which files were being downloaded and could be causing the issue. 

#### Downloading Jar files:
projects containing jar files can trigger antivirus software while being downloaded. you can either temporarily disable network antivirus protection or exclude directories containing jar files.

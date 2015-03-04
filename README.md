#CF App Downloader

##Install 
1. download binary (See Download Section)
2. **cd path/to/downloaded/binary**
3. **cf install-plugin download** (if updating run **cf uninstall-plugin download** first)
4. make sure plugin shows up in **cf plugins** 

##Download
| Mac (Darwin)  | Windows       | Linux         |
|:-------------:|:-------------:|:-------------:|
| [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/darwin/386/download) | [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/blob/master/binaries/windows/386/download.exe) | [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/386/download)
| [amd 64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/darwin/amd64/download) | [amd64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/blob/master/binaries/windows/amd64/download.exe) | [amd64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/amd64/download)
|               |               | [arm](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/arm/download) |  

##Usage
cf download <APP_NAME> [PATH] [--flags]
path and flags are optional, setting a path will start the download from the specified path instead of the app root.

##Flag Options
-i			instance
--omit		omit directory or file
--overwrite	overwrite files
--routines	max number of concurrent subroutines (default 200)
--verbose	verbose


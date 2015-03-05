CF DOWNLOAD

A Cloud Foundry cli plugin for downloading  

Install 
1. download binary (See Download Section)
2. **cd path/to/downloaded/binary**
3. Only if updating: run **cf uninstall-plugin download** first
3. **cf install-plugin download** note: may require sudo because of download file permissions
4. make sure plugin shows up in **cf plugins** 

Download

Mac: [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/darwin/386/cf-download) | [amd 64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/darwin/amd64/cf-download)   
Windows: [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/blob/master/binaries/windows/386/cf-download.exe) | [amd64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/blob/master/binaries/windows/amd64/cf-download.exe)    
Linux: [386](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/386/cf-download) | [amd64](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/amd64/cf-download) | [arm](https://github.rtp.raleigh.ibm.com/jstart/cf-download/raw/master/binaries/linux/arm/cf-download)

Usage

cf download APP_NAME [PATH] [--overwrite] [--verbose] [--omit ommited_path] [--routines max_routines] [-i instance]

path and flags are optional, setting a path will start the download from the specified path instead of the app root.
Use cf help download to see options

Notes:  
Projects usually have a enormous amount of dependencies installed by package managers, we highly reccomend not downloading these dependencies. using the --omit flag you can avoid downloading these dependencies and significantly reduce download times.

In case the download seems to be stuck, we recommend terminating and redownloading using the verbose flag. When the download stalls you can see which files were being downloaded and could be causing the issue. 
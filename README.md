# packagetools

is a utility for slackware to extract information about packages
which can be used in a shell pipeline.

## installation

with go installed and `$GOPATH/bin` in your `$PATH` just run

	go install

and you'll have packagetools installed.

## usage of packagetools

	-available
	  	list repository packages
	-list
	  	list installed packages
	-local string
	  	directory of package install logs (default "/var/log/packages")
	-prefixes string
	  	package subdirs in repository to consider (default "./patches/packages ./slackware64")
	-repo string
	  	path to repository (only CHECKSUMS.md5 and FILELIST.TXT are required) (default "/mnt/mirror/slackware/slackware64-14.2")
	-upgrade
	  	list upgradeable packages

## output format

For -available and -list the output format is a tab-seperated list of
name, version, arch, build, unix time of modification, file path

For -upgrade the output is similar, only that the first six fields are
information about the local package and the last six fields information
about the repository package.

## usage examples

this assumes the default location for repositories is valid.

### find all packages which can be upgraded, except kernel packages, and print a download url for the package and the signature

	for x in $(packagetools -upgrade | grep -v "kernel" | cut -f 12); do
		echo "http://ftp.slackware.com/pub/slackware/slackware64-14.2/$x"
		echo "http://ftp.slackware.com/pub/slackware/slackware64-14.2/$x.asc"
	done

### search a specific package in the repository and display the name, version and build

	packagetools -available | grep "bash" | cut -f 1,2,4


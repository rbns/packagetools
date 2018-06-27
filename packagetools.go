package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
)

const outputUsage = `
Output format:
For -available and -list the output format is a tab-seperated list of
name, version, arch, build, unix time of modification, file path

For -upgrade the output is similar, only that the first six fields are
information about the local package and the last six fields information
about the repository package.
`

var debug = false

//available returns a slice of infos for packages in the repository
func available(slackwareDir string, prefixes []string) ([]info, error) {
	pkgs, err := readRepository(slackwareDir, prefixes)
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}

//list locally installed packages
func list(packageLog string) ([]info, error) {
	is, err := readPackageLog(packageLog)
	if err != nil {
		return nil, err
	}

	resultIs := make([]info, len(is))
	i := 0
	for _, v := range is {
		resultIs[i] = v
		i++
	}

	return resultIs, nil
}

//upgrade matches locally installed packages to packages in the repository, selecting the newest package
//from the repository as match to a local package.
func upgrade(repo string, prefixes []string, local string) ([]upgradeInfo, error) {
	repoPkgs, err := readRepository(repo, prefixes)
	if err != nil {
		return nil, err
	}

	localPkgs, err := readPackageLog(local)
	if err != nil {
		return nil, err
	}

	out := []upgradeInfo{}
	for _, localPkg := range localPkgs {
		match := upgradeInfo{}
		for _, repoPkg := range repoPkgs {
			if repoPkg.Name == localPkg.Name {
				if repoPkg.Version != localPkg.Version || repoPkg.Build != localPkg.Build {
					match = upgradeInfo{Local: localPkg, Repo: repoPkg}
				}
				break
			}
		}

		if match == (upgradeInfo{}) {
			continue
		}

		out = append(out, match)
	}

	sort.Sort(upgradeInfoByName(out))
	return out, nil
}

func main() {
	repo := flag.String("repo", "/mnt/mirror/slackware/slackware64-14.2", "path to repository (only CHECKSUMS.md5 and FILELIST.TXT are required)")
	prefixesFlag := flag.String("prefixes", "./patches/packages ./slackware64", "package subdirs in repository to consider")
	local := flag.String("local", "/var/log/packages", "directory of package install logs")
	availableFlag := flag.Bool("available", false, "list repository packages")
	listFlag := flag.Bool("list", false, "list installed packages")
	upgradeFlag := flag.Bool("upgrade", false, "list upgradeable packages")
	flag.BoolVar(&debug, "debug", false, "debug output")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), outputUsage)
	}

	flag.Parse()

	prefixes := strings.Split(*prefixesFlag, " ")

	var err error
	var fs []fmt.Stringer

	switch {
	case *availableFlag:
		is, err2 := available(*repo, prefixes)
		err = err2
		fs = make([]fmt.Stringer, len(is))
		for i, v := range is {
			fs[i] = fmt.Stringer(v)
		}
	case *listFlag:
		is, err2 := list(*local)
		err = err2
		fs = make([]fmt.Stringer, len(is))
		for i, v := range is {
			fs[i] = fmt.Stringer(v)
		}
	case *upgradeFlag:
		is, err2 := upgrade(*repo, prefixes, *local)
		err = err2
		fs = make([]fmt.Stringer, len(is))
		for i, v := range is {
			fs[i] = fmt.Stringer(v)
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	for _, v := range fs {
		fmt.Println(v)
	}
}

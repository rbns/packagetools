package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//suffixes are common package file extensions
var suffixes = []string{".tgz", ".txz", ".tbz2", ".tar.gz", ".tar.xz", ".tar.bz2"}

//hasPrefixes tests s if it has any prefix of pre
func hasPrefixes(s string, pre ...string) bool {
	for _, v := range pre {
		if strings.HasPrefix(s, v) {
			return true
		}
	}

	return false
}

//hasSuffixes tests s if it has any suffix of suf
func hasSuffixes(s string, suf ...string) bool {
	for _, v := range suf {
		if strings.HasSuffix(s, v) {
			return true
		}
	}

	return false
}

//trimSuffixes removes any suffix in suf from s
//NB: it doesn't stop after the first removal.
func trimSuffixes(s string, suf ...string) string {
	for _, v := range suf {
		s = strings.TrimSuffix(s, v)
	}

	return s
}

//parsePackagePath parses a filename into package information.
func parsePackagePath(packagePath string) (info, error) {
	// basename of file, remove package suffixes so they don't end up in build, split at dashes
	x := strings.Split(trimSuffixes(filepath.Base(packagePath), suffixes...), "-")
	if len(x) < 3 {
		return info{}, fmt.Errorf("invalid package name with less than 3 dashes: %v", packagePath)
	}
	name := strings.Join(x[0:len(x)-3], "-")
	version := x[len(x)-3]
	arch := x[len(x)-2]
	build := x[len(x)-1]

	if name == "" || version == "" || arch == "" || build == "" {
		return info{}, fmt.Errorf("empty fields in package: %v", packagePath)
	}

	return info{Name: name, Version: version, Arch: arch, Build: build, Path: packagePath}, nil
}

//packageTimestamps reads FILELIST.TXT like contents found in slackware repositories.
//it extracts package paths and modification times. only file paths starting with one of prefixes are considered.
func packageTimestamps(filelist io.Reader, prefixes ...string) (map[string]info, error) {
	out := make(map[string]info)

	s := bufio.NewScanner(filelist)
	for s.Scan() {
		fs := strings.Fields(s.Text())
		if len(fs) != 8 {
			if debug {
				log.Println("timestamps: skipping line with field count not 8:", s.Text())
			}
			continue
		}

		// ignore non package lines (asc, txt, etc.)
		if !hasSuffixes(fs[7], suffixes...) {
			if debug {
				log.Println("timestamps: skipping line with no matching suffix in file:", s.Text())
			}
			continue
		}

		if len(prefixes) != 0 && !hasPrefixes(fs[7], prefixes...) {
			if debug {
				log.Println("timestamps: skipping line with no matching prefix in file:", s.Text())
			}
			continue
		}

		if _, ok := out[fs[7]]; ok {
			return nil, fmt.Errorf("duplicate entry in filelist: %v", fs[7])
		}

		i, err := parsePackagePath(fs[7])
		if err != nil {
			return nil, err
		}

		modTime, err := time.Parse("2006-01-02 15:04", fmt.Sprintf("%v %v", fs[5], fs[6]))
		if err != nil {
			return nil, err
		}

		i.ModTime = modTime

		out[fs[7]] = i
	}

	return out, nil
}

//packageChecksums reads CHECKSUMS.md5 like formatted contents found in slackware repositories.
//it extracts package information and the package checksum. only file paths starting with one of prefixes are considered.
func packageChecksums(checksums io.Reader, prefixes ...string) (map[string]info, error) {
	s := bufio.NewScanner(checksums)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "MD5 message digest") {
			break
		}
	}

	pkgs := make(map[string]info)
	for s.Scan() {
		fs := strings.Split(s.Text(), "  ")
		if len(fs) != 2 {
			return nil, fmt.Errorf("invalid CHECKSUMS.md5 line: %v", s.Text())
		}

		if len(prefixes) != 0 && !hasPrefixes(fs[1], prefixes...) {
			if debug {
				log.Println("checksums: skipping line with no matching prefix in file:", s.Text())
			}
			continue
		}

		// ignore non package lines (asc, txt, etc.)
		if !hasSuffixes(fs[1], suffixes...) {
			if debug {
				log.Println("checksums: skipping line with no matching suffix in file:", s.Text())
			}
			continue
		}

		i, err := parsePackagePath(fs[1])
		if err != nil {
			return nil, err
		}

		pkgs[fs[1]] = i
	}

	return pkgs, nil
}

//readRepository reads FILELIST.TXT and CHECKSUMS.md5, parsing their information into an info-slice.
func readRepository(repo string, prefixes []string) ([]info, error) {
	f, err := os.Open(filepath.Join(repo, "FILELIST.TXT"))
	if err != nil {
		return nil, err
	}

	tsPkgs, err := packageTimestamps(f, prefixes...)
	if err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	f, err = os.Open(filepath.Join(repo, "CHECKSUMS.md5"))
	if err != nil {
		return nil, err
	}

	csPkgs, err := packageChecksums(f, prefixes...)
	if err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	//merge the two maps, retaining the newest version of a package.
	newestPkgs := make(map[string]info)
	for k, v := range csPkgs {
		w, ok := tsPkgs[k]
		if !ok {
			return nil, fmt.Errorf("package in checksums not found in timestamps: %v", k)
		}

		//		mergedPkg := info{Name: v.Name, Version: v.Version, Arch: v.Arch, Build: v.Build, ModTime: w.ModTime, CheckSum: v.CheckSum, Path: v.Path}

		mergedPkg, err := mergeInfos(v, w)
		if err != nil {
			return nil, err
		}

		_, ok = newestPkgs[v.Name]
		if !ok {
			newestPkgs[v.Name] = mergedPkg
			continue
		}

		if newestPkgs[v.Name].ModTime.Before(mergedPkg.ModTime) {
			newestPkgs[v.Name] = mergedPkg
		}
	}

	out := []info{}
	for _, v := range newestPkgs {
		out = append(out, v)
	}

	sort.Sort(infoByName(out))

	return out, nil
}

type fileInfosByName []os.FileInfo

func (x fileInfosByName) Len() int           { return len(x) }
func (x fileInfosByName) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x fileInfosByName) Less(i, j int) bool { return x[i].Name() < x[j].Name() }

//readPackageLog reads a local directory of package logs returning a info-slice.
func readPackageLog(packageLog string) ([]info, error) {
	pkgs := []info{}

	f, err := os.Open(packageLog)
	if err != nil {
		return nil, err
	}

	fileInfos, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}

	// sort so we get packages alphabetically
	sort.Sort(fileInfosByName(fileInfos))

	for _, v := range fileInfos {
		i, err := parsePackagePath(v.Name())
		if err != nil {
			return nil, err
		}

		i.ModTime = v.ModTime()

		pkgs = append(pkgs, i)
	}

	return pkgs, nil
}

package main

import (
	"fmt"
	"time"
)

//info contains package information
type info struct {
	Name     string
	Version  string
	Arch     string
	Build    string
	ModTime  time.Time
	CheckSum string
	Path     string
}

func (i info) String() string {
	return fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", i.Name, i.Version, i.Arch, i.Build, i.ModTime.Unix(), i.Path)
}

func mergeInfos(a, b info) (out info, err error) {
	if a.Name != "" && b.Name == "" || a.Name == b.Name {
		out.Name = a.Name
	} else if a.Name == "" && b.Name != "" {
		out.Name = b.Name
	} else {
		return out, fmt.Errorf("Name conflict: a: %v b: %v", a.Name, b.Name)
	}

	if a.Version != "" && b.Version == "" || a.Version == b.Version {
		out.Version = a.Version
	} else if a.Version == "" && b.Version != "" {
		out.Version = b.Version
	} else {
		return out, fmt.Errorf("Version conflict: a: %v b: %v", a.Version, b.Version)
	}

	if a.Arch != "" && b.Arch == "" || a.Arch == b.Arch {
		out.Arch = a.Arch
	} else if a.Arch == "" && b.Arch != "" {
		out.Arch = b.Arch
	} else {
		return out, fmt.Errorf("Arch conflict: a: %v b: %v", a.Arch, b.Arch)
	}

	if a.Build != "" && b.Build == "" || a.Build == b.Build {
		out.Build = a.Build
	} else if a.Build == "" && b.Build != "" {
		out.Build = b.Build
	} else {
		return out, fmt.Errorf("Build conflict: a: %v b: %v", a.Build, b.Build)
	}

	if !a.ModTime.IsZero() && b.ModTime.IsZero() || a.ModTime == b.ModTime {
		out.ModTime = a.ModTime
	} else if a.ModTime.IsZero() && !b.ModTime.IsZero() {
		out.ModTime = b.ModTime
	} else {
		return out, fmt.Errorf("ModTime conflict: a: %v b: %v", a.ModTime, b.ModTime)
	}

	if a.CheckSum != "" && b.CheckSum == "" || a.CheckSum == b.CheckSum {
		out.CheckSum = a.CheckSum
	} else if a.CheckSum == "" && b.CheckSum != "" {
		out.CheckSum = b.CheckSum
	} else {
		return out, fmt.Errorf("CheckSum conflict: a: %v b: %v", a.CheckSum, b.CheckSum)
	}

	if a.Path != "" && b.Path == "" || a.Path == b.Path {
		out.Path = a.Path
	} else if a.Path == "" && b.Path != "" {
		out.Path = b.Path
	} else {
		return out, fmt.Errorf("Path conflict: a: %v b: %v", a.Path, b.Path)
	}

	return
}

type infoByName []info

func (s infoByName) Len() int           { return len(s) }
func (s infoByName) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s infoByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

//upgradeInfo pairs information about locally installed packages and packages in the repository.
type upgradeInfo struct {
	Local info
	Repo  info
}

func (u upgradeInfo) String() string {
	return fmt.Sprintf("%v\t%v", u.Local, u.Repo)
}

type upgradeInfoByName []upgradeInfo

func (s upgradeInfoByName) Len() int           { return len(s) }
func (s upgradeInfoByName) Less(i, j int) bool { return s[i].Local.Name < s[j].Local.Name }
func (s upgradeInfoByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	version "github.com/hashicorp/go-version"
)

func main() {
	gf := gradleFiles()

	gr := NewGoogleRepo()
	for _, filename := range gf {
		gr.updateGradleFile(filename)
	}
}

type Repo map[string]map[string][]string

func (g Repo) updateGradleFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	content, updated := g.updateGradleContent(string(data))
	if updated {
		ioutil.WriteFile(filename, []byte(content), 0655)
	}
}

func (g Repo) updateGradleContent(content string) (string, bool) {
	re := regexp.MustCompile("(('|\")(.*):(.*):(.*)('|\"))(\\s?//\\s*(.*))?")
	matches := re.FindAllStringSubmatch(content, -1)

	updated := false
	for _, m := range matches {
		pkg := m[3]
		module := m[4]
		version := m[5]
		semver := m[8]
		current := fmt.Sprintf("%s:%s:%s", pkg, module, version)

		if v := g.fetchVersions(pkg, module); v != nil {
			lv := latestVersion(v, version, semver)
			newest := fmt.Sprintf("%s:%s:%s", pkg, module, lv)
			if newest != current {
				updated = true
				fmt.Println(current, "->", lv)
				content = strings.Replace(content, current, newest, -1)
			}
		}
	}
	return content, updated
}

func (g Repo) fetchVersions(pkg string, module string) []string {
	group, ok := g[pkg]
	if ok && group == nil {
		g[pkg] = googleMaven(pkg)
	}
	if !ok || g[pkg][module] == nil {
		m := make(map[string][]string)
		m[module] = jcenter(pkg, module)
		g[pkg] = m
	}
	return g[pkg][module]
}

func jcenter(pkg string, module string) []string {
	n := Metadata{}
	u := fmt.Sprintf("https://jcenter.bintray.com/%s/%s/maven-metadata.xml", strings.Replace(pkg, ".", "/", -1), module)
	parseXml(u, &n)
	return n.Versions
}

// Metadata generated with https://github.com/wicast/xj2s
type Metadata struct {
	Latest      string   `xml:"versioning>latest"`
	Release     string   `xml:"versioning>release"`
	Versions    []string `xml:"versioning>versions>version"`
	LastUpdated string   `xml:"versioning>lastUpdated"`
	GroupID     string   `xml:"groupId"`
	ArtifactID  string   `xml:"artifactId"`
	Version     string   `xml:"version"`
}

func latestVersion(versions []string, currentVersion string, constraints string) string {
	if constraints != "" {
		valid, err := version.NewConstraint(constraints)
		if err != nil {
			log.Printf("constraint '%v' is invalid: %v", constraints, err)
			return currentVersion
		}
		for i := len(versions) - 1; i >= 0; i-- {
			v, err := version.NewVersion(versions[i])
			if err != nil {
				continue
			}
			if valid.Check(v) {
				return versions[i]
			}
		}
		return currentVersion
	}

	prerelease := isPrereleaseVersion(currentVersion)
	for i := len(versions) - 1; i >= 0; i-- {
		if !prerelease {
			if isPrereleaseVersion(versions[i]) {
				continue
			}
		}
		return versions[i]
	}
	return ""
}

func isPrereleaseVersion(name string) bool {
	tags := []string{"alpha", "beta", "rc", "build"}
	for _, tag := range tags {
		if strings.Contains(name, tag) {
			return true
		}
	}
	return false
}

func parseXml(url string, v interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("accept", "*/*")
	b, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer b.Body.Close()
	decoder := xml.NewDecoder(b.Body)
	return decoder.Decode(v)
}

func gradleFiles() []string {
	files := make([]string, 0)
	dir, _ := os.Getwd()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", dir, err)
			return err
		}
		if info.IsDir() && (info.Name() == "build" || info.Name() == "src") {
			return filepath.SkipDir
		}
		if info.Name() == "build.gradle" {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dir, err)
	}
	return files
}

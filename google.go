package main

import (
	"fmt"
	"strings"
)

func NewGoogleRepo() Repo {
	n := Node{}
	parseXml("https://dl.google.com/dl/android/maven2/master-index.xml", &n)

	gr := make(Repo)
	for _, nn := range n.Nodes {
		gr[nn.XMLName.Local] = nil
	}
	return gr
}

func googleMaven(pkg string) map[string][]string {
	u := fmt.Sprintf("https://dl.google.com/dl/android/maven2/%s/group-index.xml", strings.Replace(pkg, ".", "/", -1))
	n := Node{}
	parseXml(u, &n)
	m := make(map[string][]string)
	for _, nn := range n.Nodes {
		m[nn.XMLName.Local] = strings.Split(nn.Attributes[0].Value, ",")
	}
	return m
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type operatingSystems struct {
	Darwin  string `json:"darwin,omitempty"`
	Linux   string `json:"linux,omitempty"`
	Windows string `json:"windows,omitempty"`
}

type checksums struct {
	AMD64   *operatingSystems `json:"amd64,omitempty"`
	ARM     *operatingSystems `json:"arm,omitempty"`
	ARM64   *operatingSystems `json:"arm64,omitempty"`
	PPC64LE *operatingSystems `json:"ppc64le,omitempty"`
	S390X   *operatingSystems `json:"s390x,omitempty"`
	operatingSystems
}

type release struct {
	Checksums checksums `json:"checksums"`
	Name      string    `json:"name"`
}

type releases struct {
	Releases []release
}

func (r *releases) UnmarshalJSON(p []byte) error {
	return json.Unmarshal(p, &r.Releases)
}

func main() {
	releasesFile := flag.String("releases-file", "", "The path to the releases file")
	version := flag.String("version", "", "The version of minikube to create the entry for")
	flag.Parse()

	if *releasesFile == "" || *version == "" {
		fmt.Println("All flags are required and cannot be empty")
		flag.PrintDefaults()
		return
	}

	r, err := getReleases(*releasesFile)
	if err != nil {
		log.Fatal(err)
	}

	e := createBareRelease(*version)

	shaMap := getSHAMap(&e.Checksums)
	for os, archs := range shaMap {
		for arch, sumVars := range archs {
			sha, err := getSHA(os, arch)
			if err != nil {
				log.Fatal(err)
			}
			for _, sumVar := range sumVars {
				*sumVar = sha
			}

		}
	}

	r.Releases = append([]release{e}, r.Releases...)

	if err := updateJSON(*releasesFile, r); err != nil {
		log.Fatal(err)
	}
}

func getReleases(path string) (releases, error) {
	r := releases{}

	b, err := os.ReadFile(path)
	if err != nil {
		return r, fmt.Errorf("failed to read in releases file %q: %v", path, err)
	}

	if err := json.Unmarshal(b, &r); err != nil {
		return r, fmt.Errorf("failed to unmarshal releases file: %v", err)
	}

	return r, nil
}

func createBareRelease(name string) release {
	return release{
		Checksums: checksums{
			AMD64:   &operatingSystems{},
			ARM:     &operatingSystems{},
			ARM64:   &operatingSystems{},
			PPC64LE: &operatingSystems{},
			S390X:   &operatingSystems{},
		},
		Name: name,
	}
}

func getSHAMap(c *checksums) map[string]map[string][]*string {
	return map[string]map[string][]*string{
		"darwin": {
			"amd64": {&c.AMD64.Darwin, &c.Darwin},
			"arm64": {&c.ARM64.Darwin},
		},
		"linux": {
			"amd64":   {&c.AMD64.Linux, &c.Linux},
			"arm":     {&c.ARM.Linux},
			"arm64":   {&c.ARM64.Linux},
			"ppc64le": {&c.PPC64LE.Linux},
			"s390x":   {&c.S390X.Linux},
		},
		"windows": {
			"amd64": {&c.AMD64.Windows, &c.Windows},
		},
	}
}

func getSHA(operatingSystem, arch string) (string, error) {
	if operatingSystem == "windows" {
		arch += ".exe"
	}
	//filePath := fmt.Sprintf("out/minikube-%s-%s.sha256", operatingSystem, arch)
	filePath := "out/minikube-test.sha256"
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %v", filePath, err)
	}
	return string(b), nil
}

func updateJSON(path string, r releases) error {
	b, err := json.MarshalIndent(r.Releases, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal releases to JSON: %v", err)
	}

	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("failed to write JSON to file: %v", err)
	}
	return nil
}

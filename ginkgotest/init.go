package ginkgotest

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamillosantos/macchiato"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
)

var projectRoot string

func getProjectRoot() string {
	if projectRoot == "" {
		projectRoot = os.Getenv("PROJECT_ROOT")
		if dir, err := os.Getwd(); err == nil && projectRoot == "" {
			projectRoot = dir

			for {
				if _, err := os.Stat(path.Join(projectRoot, "go.sum")); err != nil {
					if projectRoot == "/" {
						projectRoot = "."
						break
					} else {
						projectRoot = path.Dir(projectRoot)
					}
				} else {
					break
				}
			}
		}
	}
	return projectRoot
}

func Init(description string, t *testing.T) {
	log.SetOutput(ginkgo.GinkgoWriter)
	gomega.RegisterFailHandler(ginkgo.Fail)

	if os.Getenv("CI") == "" {
		macchiato.RunSpecs(t, description)
	} else {
		projectRoot := getProjectRoot()
		project := filepath.Base(projectRoot)
		dir, _ := os.Getwd()
		reporterOutputDir := path.Join(projectRoot, "test-results", project, strings.Replace(dir, projectRoot, "", 1))
		os.MkdirAll(reporterOutputDir, os.ModePerm)
		junitReporter := reporters.NewJUnitReporter(path.Join(reporterOutputDir, "results.xml"))
		macchiatoReporter := macchiato.NewReporter()
		ginkgo.RunSpecsWithCustomReporters(t, description, []ginkgo.Reporter{macchiatoReporter, junitReporter})
	}
}

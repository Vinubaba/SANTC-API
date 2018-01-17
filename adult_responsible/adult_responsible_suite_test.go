package adult_responsible_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdultResponsible(t *testing.T) {
	RegisterFailHandler(Fail)
	initDb()
	defer deleteDb()
	RunSpecs(t, "AdultResponsible Suite")
}

func deleteDb() {
	if err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-c", "drop database test_teddycare").Run(); err != nil {
		log.Fatal(err.Error())
	}
}

func initDb() {
	if err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-c", "create database test_teddycare").Run(); err != nil {
		log.Fatal("failed to create database:" + err.Error())
	}

	if err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-c", "grant all privileges on database test_teddycare to postgres").Run(); err != nil {
		log.Fatal("failed to grant privileges:" + err.Error())
	}

	visit := func(files *[]string) filepath.WalkFunc {
		return func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			*files = append(*files, path)
			return nil
		}
	}

	var files []string

	root := `C:\Users\arthur\gocode\src\arthurgustin.fr\teddycare\sql\`
	err := filepath.Walk(root, visit(&files))
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.Contains(file, "up") {
			cmd := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-d", "test_teddycare", "-a", "-f", fmt.Sprintf("%s", file))
			err = cmd.Run()
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}

}

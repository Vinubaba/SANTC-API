package users_test

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

func TestUsers(t *testing.T) {
	RegisterFailHandler(Fail)
	initDb()
	defer deleteDb()
	RunSpecs(t, "Users Suite")
}

func deleteDb() {
	if out, err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-c", "drop database test_teddycare").Output(); err != nil {
		log.Println(string(out))
		log.Fatal(err.Error())
	}
}

func initDb() {
	if out, err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-c", "create database test_teddycare").Output(); err != nil {
		log.Println(string(out))
		//log.Fatal("failed to create database:" + err.Error())
	}

	if out, err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-c", "grant all privileges on database test_teddycare to postgres").Output(); err != nil {
		log.Fatal("failed to grant privileges:" + string(out))
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

	root := os.Getenv("TEDDYCARE_SQL_DIR")
	if root == "" {
		root = `C:\Users\arthur\gocode\src\github.com\DigitalFrameworksLLC\teddycare\sql`
		log.Println("please set env TEDDYCARE_SQL_DIR")
		log.Println("default to " + root)
	}
	log.Println("will use " + root)

	err := filepath.Walk(root, visit(&files))
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.Contains(file, "up") {
			out, err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-d", "test_teddycare", "-a", "-f", fmt.Sprintf("%s", file)).Output()
			if err != nil {
				log.Print(string(out))
				log.Fatal(err.Error())
			}
		}
	}
}

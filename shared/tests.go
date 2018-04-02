package shared

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func DeleteDb() {
	if out, err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-c", "drop database test_teddycare").Output(); err != nil {
		log.Println(string(out))
		log.Fatal(err.Error())
	}
}

func InitDb() {
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

	root := getSqlDirPath(true)

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

func getSqlDirPath(verbose bool) string {
	root := os.Getenv("TEDDYCARE_SQL_DIR")
	if root == "" {
		root = `C:\Users\arthur\gocode\src\github.com\Vinubaba\SANTC-API\sql`
		if verbose {
			log.Println("please set env TEDDYCARE_SQL_DIR")
			log.Println("default to " + root)
		}
	}
	if verbose {
		log.Println("will use " + root)
	}
	return root
}

func SetDbInitialState() {
	root := getSqlDirPath(false)
	out, err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-d", "test_teddycare", "-a", "-f", path.Join(root, "test_initial_state.sql")).Output()
	if err != nil {
		log.Print(string(out))
		log.Fatal(err.Error())
		panic(err)
	}
}

func NewDbInstance(verbose bool, logger ...interface {
	Print(v ...interface{})
}) *gorm.DB {
	var err error
	var db *gorm.DB
	connectString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		"localhost",
		"5432",
		"postgres",
		"postgres",
		"test_teddycare")
	db, err = gorm.Open("postgres", connectString)
	if err != nil {
		panic(err)
	}
	db.LogMode(verbose)
	if len(logger) > 0 {
		db.SetLogger(logger[0])
	}
	return db
}

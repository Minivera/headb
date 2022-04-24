package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/go-jet/jet/v2/generator/postgres"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var databaseInfoExtractor = regexp.MustCompile(
	`postgresql://(?P<username>[^:]*):(?P<password>[^@]*)@(?P<host>[^:]*):(?P<port>[^/]*)/(?P<database>[^?]*)\?.*`,
)

func main() {
	service := flag.String("service", "", "Which service to generate the database for")

	flag.Parse()

	if service == nil {
		log.Fatal("Please provide the service name")
		return
	}

	fullPath, err := exec.LookPath("encore")
	if err != nil {
		log.WithError(err).Fatal("Could not find path to encore binary")
		return
	}

	out, err := exec.Command(fullPath, "db", "conn-uri", *service).Output()
	if err != nil {
		log.WithError(err).Fatalf("Could not execute encore command, out: %s", out)
		return
	}

	match := databaseInfoExtractor.FindStringSubmatch(string(out))
	result := make(map[string]string)
	for i, name := range databaseInfoExtractor.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	port, err := strconv.Atoi(result["port"])
	if err != nil {
		log.WithError(err).Fatal("Could not convert port string to integer")
		return
	}

	err = postgres.Generate("./generated", postgres.DBConnection{
		Host:       result["host"],
		Port:       port,
		User:       result["username"],
		Password:   result["password"],
		DBName:     result["database"],
		SchemaName: "public",
		SslMode:    "disable",
	})
	if err != nil {
		log.WithError(err).Fatal("Could not execute jet database model generation")
		return
	}

	log.Info("Replacing all types for uuids to encore uuids")
	err = filepath.Walk("./generated", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileExtension := filepath.Ext(path)

		if info.IsDir() || fileExtension != ".go" {
			log.Infof("Ignored file %s because it was a directory or not a go file", path)
			return nil
		}

		input, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		output := bytes.Replace(input, []byte("github.com/google/uuid"), []byte("encore.dev/types/uuid"), -1)

		if err = ioutil.WriteFile(path, output, info.Mode()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Fatal("Could not list generated files or replace their content")
		return
	}

	log.Infof("Models generated for service %s", *service)
}

package main

import (
	"flag"
	"os/exec"
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

	log.Infof("Models generated for service %s", *service)
}

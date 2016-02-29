package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/codegangsta/cli"
)

const (
	ENV_SQL_DATASOURCE = "ATC_SQL_DATASOURCE"
	ENV_SQL_DRIVER     = "ATC_SQL_DRIVER"
)

var atcFlags = []cli.Flag{
	cli.BoolFlag{Name: "dev", Usage: "dev mode; lax security", EnvVar: "ATC_DEV"},
	cli.StringFlag{Name: "callbacksURL", Value: "http://127.0.0.1:8080", Usage: "URL used for callbacks to reach the ATCD (excluding basic auth)", EnvVar: "ATC_CALLBACK_URL"},
	cli.StringFlag{Name: "checkInterval", Value: "1m0s", Usage: "interval on which to poll for new versions of resources", EnvVar: "ATC_CHECK_INTERVAL"},
	cli.StringFlag{Name: "cliDownloadsDir", Value: "", Usage: "directory containing CLI binaries to serve", EnvVar: "ATC_CLI_DOWNLOADS_DIR"},
	cli.StringFlag{Name: "httpUsername", Value: "", Usage: "basic auth username for the server", EnvVar: "ATC_USERNAME"},
	cli.StringFlag{Name: "httpPassword", Value: "", Usage: "basic auth password for the server", EnvVar: "ATC_PASSWORD"},
	cli.StringFlag{Name: "sqlDataSource", Value: "postgres://127.0.0.1:5432/atc?sslmode=disable", Usage: "database/sql data source configuration string", EnvVar: ENV_SQL_DATASOURCE},
	cli.StringFlag{Name: "sqlDriver", Value: "postgres", Usage: "database/sql driver name", EnvVar: ENV_SQL_DRIVER},
	cli.StringFlag{Name: "public", Value: "web/public", Usage: "path to directory containing public resources (javascript, css, etc.)", EnvVar: "ATC_PUBLIC"},
	cli.StringFlag{Name: "templates", Value: "web/templates", Usage: "path to directory containing the html templates", EnvVar: "ATC_TEMPLATES"},
	cli.IntFlag{Name: "webListenPort", Value: 8080, Usage: "port for the web server to listen on", EnvVar: "PORT"},
}

func main() {
	bindLinkedDockerDataSource()

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "atc", Value: "atc", Usage: "path to atc command", EnvVar: "ATC_ATC"},
	}
	app.Commands = []cli.Command{
		{
			Name:        "start",
			Description: "start the atc server",
			Flags:       atcFlags,
			Action:      Start,
		},
	}
	app.Run(os.Args)
}

func bindLinkedDockerDataSource() {
	if os.Getenv(ENV_SQL_DATASOURCE) != "" {
		return
	}

	log.Println("scanning for linked docker container named 'db'")

	user := "postgres"
	if v := os.Getenv("DB_ENV_POSTGRES_USER"); v != "" {
		user = v
	}

	password := os.Getenv("DB_ENV_POSTGRES_PASSWORD")
	ipAddr := os.Getenv("DB_PORT_5432_TCP_ADDR")

	if user != "" && password != "" && ipAddr != "" {
		log.Printf("found container, db.  updating %s\n", ENV_SQL_DATASOURCE)
		dataSource := fmt.Sprintf("postgres://%s:%s@%s:5432/atc?sslmode=disable", user, password, ipAddr)
		os.Setenv(ENV_SQL_DATASOURCE, dataSource)
	}
}

func makeArgs(c *cli.Context) []string {
	args := []string{}

	for _, flag := range atcFlags {
		var name string

		switch v := flag.(type) {
		case cli.StringFlag:
			name = v.Name
		case cli.IntFlag:
			name = v.Name
		case cli.BoolFlag:
			name = v.Name
			if c.Bool(name) {
				args = append(args, fmt.Sprintf("-%s", name))
				continue
			}
		}

		value := c.String(name)
		if value != "" {
			args = append(args, fmt.Sprintf("-%s=%s", name, value))
		}
	}

	return args

}

func Start(c *cli.Context) {
	args := makeArgs(c)

	atc := c.GlobalString("atc")
	log.Printf("=> %s %s", atc, strings.Join(args, " "))

	cmd := exec.Command(atc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

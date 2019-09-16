package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jakubknejzlik/golang-sandbox-project/src"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "golang-sandbox-project"
	app.Usage = "just for example purpose"
	app.Version = "0.0.0"

	app.Commands = []cli.Command{
		startCmd,
		migrateCmd,
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

var startCmd = cli.Command{
	Name:  "start",
	Usage: "start api server",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "p,port",
			Usage:  "Port to listen to",
			Value:  "80",
			EnvVar: "PORT",
		},
	},
	Action: func(ctx *cli.Context) error {
		port := ctx.String("port")
		if err := startServer(port); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	},
}

var migrateCmd = cli.Command{
	Name:  "migrate",
	Usage: "migrate schema database",
	Action: func(ctx *cli.Context) error {
		fmt.Println("starting migration")
		if err := automigrate(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		fmt.Println("migration complete")
		return nil
	},
}

func automigrate() error {
	db := src.NewDBFromEnvVars()
	defer db.Close()
	return db.AutoMigrate(src.Message{}).Error
}

func startServer(port string) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	db := src.NewDBFromEnvVars()
	defer db.Close()

	mux := src.GetHTTPServeMux(db)

	mux.HandleFunc("/healthcheck", func(res http.ResponseWriter, req *http.Request) {
		if err := db.DB().Ping(); err != nil {
			res.WriteHeader(400)
			res.Write([]byte("ERROR: " + err.Error()))
			return
		}
		res.WriteHeader(200)
		res.Write([]byte("OK"))
	})

	h := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		log.Printf("server running on http://localhost:%s/", port)
		log.Fatal(h.ListenAndServe())
	}()

	<-stop

	log.Println("\nShutting down the server...")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	err := h.Shutdown(ctx)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	log.Println("Server gracefully stopped")

	err = db.Close()
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	log.Println("Database connection closed")

	return nil
}

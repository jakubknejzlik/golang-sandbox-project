package main

import (
	"context"
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
		cli.StringFlag{
			Name:   "ssl_cert_path",
			Usage:  "Path to ssl certificate",
			EnvVar: "SSL_CERT_PATH",
		},
		cli.StringFlag{
			Name:   "ssl_key_path",
			Usage:  "Path to ssl certificate key",
			EnvVar: "SSL_KEY_PATH",
		},
	},
	Action: func(ctx *cli.Context) error {
		port := ctx.String("port")
		sslCertPath := ctx.String("ssl_cert_path")
		sslKeyPath := ctx.String("ssl_key_path")

		options := startServerOptions{
			port:        port,
			sslCertPath: sslCertPath,
			sslKeyPath:  sslKeyPath,
		}
		if err := startServer(options); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		return nil
	},
}

var migrateCmd = cli.Command{
	Name:  "migrate",
	Usage: "migrate schema database",
	Action: func(ctx *cli.Context) error {
		log.Println("starting migration")
		if err := automigrate(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		log.Println("migration complete")
		return nil
	},
}

func automigrate() error {
	db := src.NewDBFromEnvVars()
	defer db.Close()
	return db.AutoMigrate(src.Message{}).Error
}

type startServerOptions struct {
	port        string
	sslCertPath string
	sslKeyPath  string
}

func startServer(opts startServerOptions) error {
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

	h := &http.Server{Addr: ":" + opts.port, Handler: mux}

	go func() {
		log.Printf("server running on http://localhost:%s/", opts.port)
		if opts.sslCertPath != "" && opts.sslKeyPath != "" {
			log.Fatal(h.ListenAndServeTLS(opts.sslCertPath, opts.sslKeyPath))
		} else {
			log.Fatal(h.ListenAndServe())
		}
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

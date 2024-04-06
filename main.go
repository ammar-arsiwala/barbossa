package main

import (
	"flag"
	"log"
	"os"
	"runtime/trace"

	"github.com/gofiber/fiber/v2"
	route "github.com/keshavchand/barbossa/routes"
	"github.com/keshavchand/barbossa/service"
)

func main() {
	traceFileName := "-"
	flag.StringVar(&traceFileName, "trace", "-", "trace output file")
	flag.Parse()

	if traceFileName != "-" {
		traceOut, err := os.OpenFile(traceFileName, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer traceOut.Close()

		if err = trace.Start(traceOut); err != nil {
			log.Fatal(err)
		}

		defer trace.Stop()
	}

	cli, err := service.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{})
	routes := []route.Route{
		route.Status,
		route.Startup,
		route.Shutdown,
		route.Connect,
		route.Partition,
	}

	for _, route := range routes {
		route(app, cli)
	}

	log.Println("Server up and running")
	if err := app.Listen(":8080"); err != nil {
		log.Println(err)
	}
}

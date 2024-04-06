package route

import (
	"runtime/trace"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"github.com/keshavchand/barbossa/service"
)

type Route func(*fiber.App, *client.Client)

func Status(app *fiber.App, cli *client.Client) {
	app.Get("/status", func(c *fiber.Ctx) error {
		trace.Logf(c.Context(), "HTTP Request", "/status")
		return nil
	})
}

var (
	naiveEndpointStorage = service.NewNaiveEndpointStorage()
)

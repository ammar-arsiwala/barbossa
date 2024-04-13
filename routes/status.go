package route

import (
	"fmt"
	"runtime/trace"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"github.com/keshavchand/barbossa/models"
	"github.com/keshavchand/barbossa/service"
)

type Route func(*fiber.App, *client.Client)

func Status(app *fiber.App, cli *client.Client) {
	// swagger:route /status Status
	// ---
	// summary: Get the status of the service.
	// description: Get the status of the service.
	// responses:
	//   200: StatusResponse
	app.Get("/status", func(c *fiber.Ctx) error {
		trace.Logf(c.Context(), "HTTP Request", "/status")
		name := c.Query("name")

		if name == "" {
			c.Status(400)
			fmt.Fprintf(c, "Error: name should not be empty")
			return nil
		}

		svc := service.NewStatus(models.StatusRequest{
			Name: name,
		})

		resp, err := svc.GetStatus(c.Context(), cli)
		if err != nil {
			c.Status(500)
			fmt.Fprintf(c, "Error %s", err.Error())
			return err
		}

		return c.JSON(resp)
	})
}

var (
	naiveEndpointStorage = service.NewNaiveEndpointStorage()
)

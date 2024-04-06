package route

import (
	"encoding/json"
	"fmt"
	"runtime/trace"

	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"github.com/keshavchand/barbossa/models"
	"github.com/keshavchand/barbossa/service"
)

func Shutdown(app *fiber.App, cli *client.Client) {
	app.Post("/shutdown", func(c *fiber.Ctx) error {
		trace.Logf(c.Context(), "HTTP Request", "/shutdown")
		var req models.ShutdownRequest
		err := json.Unmarshal(c.Body(), &req)
		if err != nil {
			c.Status(500)
			fmt.Fprintf(c, "Error %s", err.Error())
			return err
		}

		if err := req.Verify(); err != nil {
			c.Status(400)
			fmt.Fprintf(c, "Error %s", err.Error())
			return err
		}

		cmds := service.NewShutdown(req)
		for idx, cmd := range cmds {
			if err := cmd.Perform(c.Context(), cli); err != nil {
				c.Status(500)
				fmt.Fprintf(c, "Error (%d) %s", idx, err.Error())
				return err
			}
		}

		c.Status(202)
		return nil
	})
}

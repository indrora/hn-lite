package reload

import (
	_ "embed"
	"fmt"
	"html/template"

	fiberHtml "github.com/gofiber/template/html/v2"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var launchUUID uuid.UUID

func init() {
	var err error
	launchUUID, err = uuid.NewUUID()
	if err != nil {
		panic(err)
	}
}

//go:embed reload.js
var reloadjs []byte

func Attach(prefix string, app *fiber.App, tt *fiberHtml.Engine) {

	group := app.Group(prefix)

	group.Get("/uuid", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"launchUUID": launchUUID,
		})
	})
	group.Get("/rl.js", func(c *fiber.Ctx) error {
		c.Response().Header.Add("content-type", "text/javascript")
		return c.Send(reloadjs)
	})

	tt.AddFunc("reloadjs", func() template.HTML {
		return template.HTML(fmt.Sprintf("<script type='text/javascript' href='%s/lr.js' />", prefix))
	})
}

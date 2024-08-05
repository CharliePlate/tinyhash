package view

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/charlieplate/TinyHash/ui/internal/component"
	"github.com/labstack/echo/v4"
)

func ServeStaticFiles(c echo.Context) error {
	filePath := c.Request().URL.Path[len("/static/"):]
	fullPath := filepath.Join("../", "static", filePath)

	if strings.HasSuffix(c.Request().URL.Path, ".wasm") {
		c.Response().Header().Add("Content-Type", "application/wasm")
	}

	http.ServeFile(c.Response(), c.Request(), fullPath)

	return nil
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		log.Println("Error rendering components:", err)
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}

func Home(ctx echo.Context) error {
	log.Println(":) Someone tried to go home")

	return Render(ctx, http.StatusOK, component.Home("Hello World"))
}

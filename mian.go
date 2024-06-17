package main

import (
	"flag"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"net/http"
	"webhook-transfer/separation"
)

func main() {
	listenAddr := flag.String("port", ":80", "http listen addr")
	flag.Parse()
	seps := separation.New()
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.POST("/transfer/:name", func(c echo.Context) error {
		name := c.Param("name")
		key := c.QueryParam("key")
		if len(key) == 0 {
			return c.String(http.StatusBadRequest, "Key is empty")
		}
		if sep, ok := seps[name]; ok {
			b, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return c.String(http.StatusInternalServerError,
					fmt.Sprintf("Read request body err: %s", err.Error()))
			}
			err = sep.ConvertAndSend(b, key)
			if err != nil {
				return c.String(http.StatusInternalServerError,
					fmt.Sprintf("Convert and send err: %s", err.Error()))
			}
		} else {
			return c.String(http.StatusInternalServerError,
				fmt.Sprintf("Unsupported hook type(%s)", name))
		}
		return c.String(http.StatusOK, "Success")
	})
	e.Logger.Fatal(e.Start(*listenAddr))
}

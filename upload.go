package main

import (
	"fmt"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/access"
	"github.com/go-ozzo/ozzo-routing/content"
	"github.com/go-ozzo/ozzo-routing/fault"
	"github.com/go-ozzo/ozzo-routing/slash"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {

	router := routing.New()

	router.Use(
		access.Logger(log.Printf),
		slash.Remover(http.StatusMovedPermanently),
		fault.Recovery(log.Printf),
	)

	api := router.Group("/v1")
	api.Use(
		content.TypeNegotiator(content.JSON),
	)

	api.Post("/upload", func(c *routing.Context) error {

		c.Request.ParseMultipartForm(32 << 20)
		file, handler, err := c.Request.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return nil
		}
		defer file.Close()
		f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		defer f.Close()
		io.Copy(f, file)

		return c.Write(handler.Filename)
	})

	http.Handle("/", router)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

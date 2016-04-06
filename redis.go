package main

import (
  "os"
  "fmt"
  "log"
  "net/http"
  "encoding/json"
  "gopkg.in/redis.v3"
  "github.com/go-ozzo/ozzo-routing"
  "github.com/go-ozzo/ozzo-routing/fault"
  "github.com/go-ozzo/ozzo-routing/slash"
  "github.com/go-ozzo/ozzo-routing/access"
  "github.com/go-ozzo/ozzo-routing/content"
)

func main() {

	red := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := red.Ping().Result()
	if pong != "PONG" {
		fmt.Printf("Failed to connect to Redis", err)
		os.Exit(1)
	}

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

	api.Get(`/hgetall/<id:\D+>`, func(c *routing.Context) error {

		redHash, err := red.HGetAllMap(c.Param("id")).Result()
		 if err == redis.Nil {
			return c.Write("None")
		}

		return c.Write(redHash)
	})

	api.Post(`/hmset/<id:\D+>`, func(c *routing.Context) error {

		var req json.RawMessage
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			return err
		}

		cMap := make(map[string]string)

		e := json.Unmarshal(req, &cMap)
		if e != nil {
			panic(e)
		}

		for keys, vals := range cMap {
			red.HMSet(c.Param("id"), keys, vals)
		}

		return c.Write(cMap)
	})

	api.Put(`/set/<id:\D+>`, func(c *routing.Context) error {
		err := red.Set("data", c.Param("id"), 0).Err()
		if err != nil {
			panic(err)
		}

		dataGet, err := red.Get("data").Result()
		if err == redis.Nil {
			return c.Write("None")
		}

		return c.Write(dataGet)
	})

	api.Get(`/get/<id:\D+>`, func(c *routing.Context) error {
		val, err := red.Get(c.Param("id")).Result()
		if err == redis.Nil {
			return c.Write("None")
		}

		return c.Write(val)
	})

	http.Handle("/", router)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

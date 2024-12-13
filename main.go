package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	//logger
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/valyala/fasthttp"
)

// Message represents the structure of the JSON data to send.
type Message struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

func main() {
	// Fiber instance
	app := fiber.New()

	// CORS for external resources
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Cache-Control",
		AllowCredentials: false,
	}))
	app.Use(logger.New())

	app.Get("/sse/:id", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		id := c.Params("id")

		c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			var i int
			for {
				i++
				msg := Message{
					ID:        id,
					Timestamp: time.Now(),
					Message:   fmt.Sprintf("The time is %v", time.Now()),
				}

				jsonData, err := json.Marshal(msg)
				if err != nil {
					fmt.Printf("Error while marshalling JSON: %v\n", err)
					break
				}

				fmt.Fprintf(w, "data: %s\n\n", jsonData)
				// fmt.Println(string(jsonData))

				err = w.Flush()
				if err != nil {
					// Close the connection if flushing fails (e.g., client disconnects)
					fmt.Printf("Error while flushing: %v of %v Closing http connection.\n", err, msg.ID)
					break
				}
				time.Sleep(2 * time.Second)
			}
		}))

		return nil
	})

	// Start server
	log.Fatal(app.Listen(":3000"))
}

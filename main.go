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
	Status    bool      `json:"status"`
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
			msgChan := make(chan Message)
			done := make(chan struct{})

			msg := Message{
				ID:        id,
				Timestamp: time.Now().UTC().Add(7 * time.Hour),
				Message:   fmt.Sprintf("wait queue"),
				Status:    false,
			}

			// Timeout timer for 5 minutes
			timeout := time.After(5 * time.Minute)

			// Goroutine สำหรับ processData
			go func() {
				time.Sleep(10 * time.Second) // ตัวอย่างดีเลย์
				msg = processData(msg)
				msgChan <- msg
				close(done)
			}()

			for {
				select {
				case updatedMsg := <-msgChan:
					jsonData, err := json.Marshal(updatedMsg)
					if err != nil {
						fmt.Printf("Error while marshalling JSON: %v\n", err)
						break
					}

					fmt.Fprintf(w, "data: %s\n\n", jsonData)
					err = w.Flush()
					if err != nil {
						fmt.Printf("Error while flushing: %v. Closing HTTP connection.\n", err)
						return
					}

					if updatedMsg.Status {
						fmt.Fprint(w, "data: close\n\n")
						_ = w.Flush()
						return
					}

				case <-time.After(2 * time.Second):
					// ส่งข้อความสถานะล่าสุดทุก 2 วินาที
					msg.Timestamp = time.Now()
					msg.Message = fmt.Sprintf("The time is %v", time.Now())

					jsonData, err := json.Marshal(msg)
					if err != nil {
						fmt.Printf("Error while marshalling JSON: %v\n", err)
						break
					}

					fmt.Fprintf(w, "data: %s\n\n", jsonData)
					err = w.Flush()
					if err != nil {
						fmt.Printf("Error while flushing: %v. Closing HTTP connection.\n", err)
						return
					}

				case <-timeout:
					// ปิดการเชื่อมต่อเมื่อครบ 5 นาที
					fmt.Fprint(w, "data: close\n\n")
					_ = w.Flush()
					fmt.Println("Connection timed out. Closing connection.")
					return

				case <-done:
					return
				}
			}
		}))

		return nil
	})

	// Start server
	log.Fatal(app.Listen(":3000"))
}

func processData(msg Message) Message {
	msg.Status = true
	msg.Message = "success"
	return msg
}

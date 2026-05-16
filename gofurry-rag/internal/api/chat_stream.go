package api

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofurry/gofurry-rag/internal/db"
	"github.com/gofurry/gofurry-rag/internal/service"
	"github.com/gofiber/fiber/v3"
)

func (s *Server) chatStream(c fiber.Ctx) error {
	if s.service == nil {
		return fail(c, fiber.ErrServiceUnavailable)
	}

	var req service.QueryRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return fail(c, err)
	}
	if err := s.requireDetailedQueryAdmin(c, req.IncludeDetails); err != nil {
		return fail(c, err)
	}

	c.Set(fiber.HeaderContentType, "text/event-stream; charset=utf-8")
	c.Set(fiber.HeaderCacheControl, "no-cache, no-transform")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	return c.SendStreamWriter(func(w *bufio.Writer) {
		ctx := c.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		writeEvent := func(event string, payload any) error {
			if event != "" {
				if _, err := fmt.Fprintf(w, "event: %s\n", event); err != nil {
					return err
				}
			}
			if payload == nil {
				if _, err := fmt.Fprint(w, "data: {}\n\n"); err != nil {
					return err
				}
			} else {
				data, err := json.Marshal(payload)
				if err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
					return err
				}
			}
			return w.Flush()
		}

		callbacks := service.QueryCallbacks{
			Status: func(stage, message string) error {
				return writeEvent("status", fiber.Map{
					"stage":   stage,
					"message": message,
				})
			},
			Sources: func(sources []db.Source) error {
				return writeEvent("sources", fiber.Map{
					"sources": sources,
				})
			},
			Delta: func(text string) error {
				return writeEvent("delta", fiber.Map{
					"text": text,
				})
			},
		}

		response, err := s.service.StreamQuery(ctx, req, callbacks)
		if err != nil {
			status, _ := errorStatus(err)
			c.Status(status)
			_ = writeEvent("error", fiber.Map{
				"message": err.Error(),
			})
			return
		}
		_ = writeEvent("done", response)
	})
}

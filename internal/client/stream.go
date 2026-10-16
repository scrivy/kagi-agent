package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"charm.land/log/v2"
)

type StreamEvent struct {
	Data map[string]any
	Err  error
}

func (c *Client) GetTurnStream(ctx context.Context, turnID string) (chan StreamEvent, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		baseUri+"/api/v2/turns/"+turnID+"/stream",
		nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest(): %w", err)
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("c.c.Do(): %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("http status: %s", resp.Status)
	}

	eventChan := make(chan StreamEvent, 100)

	go func() {
		defer resp.Body.Close()
		defer close(eventChan)

		c.logger.Debug("Streaming events:")
		logLevel := c.logger.GetLevel()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1 MiB max token
		for scanner.Scan() {
			line := scanner.Text()

			if len(line) < 6 {
				continue
			}

			var data []byte
			switch line[0:4] {
			case "id: ":
				// id: 1790007828360-0
				continue
			case "data":
				data = []byte(line[6:])
			default:
				eventChan <- StreamEvent{
					Err: fmt.Errorf("unknown turn stream preamble: %s\n", line),
				}
				continue
			}

			if logLevel == log.DebugLevel {
				// print indented json event
				var indented bytes.Buffer
				err = json.Indent(&indented, data, "", "  ")
				if err != nil {
					c.logger.Errorf("json.Indent(): %s", err)
					continue
				}

				c.logger.Debug(indented.String() + "\n")
			}

			//	var event models.TurnStreamEvent
			var resp map[string]any
			err = json.Unmarshal(data, &resp)
			if err != nil {
				eventChan <- StreamEvent{
					Err: fmt.Errorf("json.Unmarshal(): %w, line: %s", err, data),
				}
				continue
			}

			eventChan <- StreamEvent{
				Data: resp,
			}
		}
		if err := scanner.Err(); err != nil {
			eventChan <- StreamEvent{
				Err: fmt.Errorf("scanner.Err(): %w", err),
			}
		}
	}()

	return eventChan, nil
}

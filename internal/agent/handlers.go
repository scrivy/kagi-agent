package agent

import (
	"context"
	"fmt"

	"github.com/scrivy/kagi-agent/internal/client"
)

func (a *Agent) HandleTurnStreamEvents(ctx context.Context, eventChan <-chan client.StreamEvent) error {
	for event := range eventChan {
		if event.Err != nil {
			a.logger.Errorf("event.Err: %s", event.Err)
			continue
		}

		objAny, exists := event.Data["event"]
		if exists {
			a.output(".")

			obj, ok := objAny.(map[string]any)
			if !ok {
				a.logger.Debugf("event is not an object. schema not recognized: %+v", event.Data)
				continue
			}

			sourceAny, exists := obj["source"]
			if !exists {
				a.logger.Debugf("source does not exist. schema not recognized: %+v", event.Data)
				continue
			}
			source, ok := sourceAny.(string)
			if !ok {
				a.logger.Debugf("source not a string. schema not recognized: %+v", event.Data)
				continue
			}

			eventTypeAny, exists := obj["type"]
			if !exists {
				a.logger.Debugf("type does not exist. schema not recognized: %+v", event.Data)
				continue
			}
			eventType, ok := eventTypeAny.(string)
			if !ok {
				a.logger.Debugf("type not a string. schema not recognized: %+v", event.Data)
				continue
			}

			payload := make(map[string]any)
			payloadAny, exists := obj["payload"]
			if exists {
				payload, ok = payloadAny.(map[string]any)
				if !ok {
					a.logger.Debugf("payload not an object. schema not recognized: %+v", event.Data)
					continue
				}
			}

			switch eventType {
			case "node.delta",
				"turn.started",
				"turn.completed",
				"node.created",
				"node.completed":

				switch source {
				case "assistant":
				case "system":
					continue
				default:
					a.logger.Debugf("node.created event type source not recognized: %+v", event.Data)
					continue
				}

				contentAny, exists := payload["content"]
				if !exists {
					// a.logger.Debugf("content does not exist. schema not recognized: %+v", event.Data)
					// happens, just a completed status update
					continue
				}
				content, ok := contentAny.(string)
				if !ok {
					a.logger.Debugf("content not a string. schema not recognized: %+v", event.Data)
					continue
				}

				a.output(fmt.Sprintf("\n\n%s\n\n", content))

				if _, err := a.parseAndSaveFiles(content); err != nil {
					return fmt.Errorf("a.parseAndSaveDiffs(): %w", err)
				}

				/* TODO: remove
				if err := a.applyDiffs(ctx, diffs); err != nil {
					return fmt.Errorf("applyDiffs(): %w", err)
				}
				*/

				continue
			case "source.read":

				continue

			default:
				a.logger.Debugf("event type not implemented: %+v", event.Data)
				continue
			}

			continue
		}

		titleAny, exists := event.Data["title"]
		// {"title": "Simple Addition Calculation"}
		if exists {
			title, ok := titleAny.(string)
			if !ok {
				a.logger.Debugf("title not a string: %+v", titleAny)
				continue
			}

			a.state.title = title
			a.update(func() {
				fmt.Fprintf(a.ui.header, "title: %s\n", title)
			})

			continue
		}

	}

	return nil
}

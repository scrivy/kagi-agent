package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"charm.land/log/v2"
	"github.com/rivo/tview"

	"github.com/scrivy/kagi-agent/internal/client"
	"github.com/scrivy/kagi-agent/internal/client/models"
	"github.com/scrivy/kagi-agent/internal/config"
)

type Agent struct {
	logger *log.Logger
	cfg    *config.Config
	cache  *config.Cache
	client *client.Client

	state state
	ui    ui

	input chan string

	debug bool
}

type state struct {
	init           *models.GetInitResponse
	filePathsInCwd []string
	model          models.Model
	title          string
}

type ui struct {
	app    *tview.Application
	pages  *tview.Pages
	header *tview.TextView
	output *tview.TextView
	input  *tview.InputField
}

func New() *Agent {
	a := Agent{
		ui: ui{
			app:    tview.NewApplication(),
			pages:  tview.NewPages(),
			header: tview.NewTextView(),
			output: tview.NewTextView(),
			input:  tview.NewInputField(),
		},
		input: make(chan string, 1),
	}

	a.logger = log.New(&a)

	return &a
}

func (a *Agent) Run() error {
	a.initUi()

	ctx, cancelCause := context.WithCancelCause(context.Background())

	// start UI
	go func() {
		cancelCause(a.ui.app.Run())
	}()
	defer a.ui.app.Stop()

	// loop until configuration and api init succeeds
	var err error
	for {
		err = a.loadConfiguration(ctx)
		if err != nil {
			return fmt.Errorf("a.loadConfiguration(): %w", err)
		}

		a.logger.SetLevel(a.cfg.LogLevel)
		a.logger.Debugf("config: %+v", *a.cfg)

		a.client, err = client.New(a.logger, a.cfg.SessionToken)
		if err != nil {
			return fmt.Errorf("httpclient.New(): %w", err)
		}

		a.update(func() {
			a.ui.input.SetLabel("authenticating with kagi...")
		})

		a.state.init, err = a.client.GetInit(ctx)
		if err != nil {
			if errors.Is(err, client.ErrUnauthorized) {
				if err := config.DeleteConfig(); err != nil {
					return fmt.Errorf("config.DeleteConfig(): %w", err)
				}

				if err := config.DeleteCache(); err != nil {
					return fmt.Errorf("config.DeleteCache(): %w", err)
				}

				continue
			}

			return fmt.Errorf("client.GetInit(): %w", err)
		}

		a.update(func() {
			a.ui.input.SetLabel("")
			a.ui.output.SetText("Successfully authenticated with Kagi.\n\n")
		})

		break
	}

	if a.cfg.LogLevel == log.DebugLevel {
		a.debug = true

		go func() {
			a.logger.Error(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// thread loop
	for {
		model := a.cache.LastModel
		if a.cfg.AlwaysAskForModel || model == "" {
			a.cache.LastModel, err = a.selectModel(ctx)
			if err != nil {
				return fmt.Errorf("a.selectModel(): %w", err)
			}

			model = a.cache.LastModel
		}

		a.update(func() {
			a.ui.header.SetText(fmt.Sprintf("model: %s\n", model))
		})

		var found bool
		for _, details := range a.state.init.Models.Models {
			if details.ID == model {
				found = true
				a.state.model = details
				break
			}
		}
		if !found { // reset model
			a.cache.LastModel = ""
			continue
		}

		// ask for reasoning effort
		if a.cfg.AlwaysAskForReasoningEffort {
			err = a.selectReasoningEffort(ctx, model)
			if err != nil {
				return fmt.Errorf("a.selectReasoningEffort(): %w", err)
			}
		}

		a.checkReasoningEffort()

		if err := a.cache.Save(); err != nil {
			return fmt.Errorf("a.cache.Save(): %w", err)
		}

		var attachmentUUIDs []string

		// TODO: put in it's own function: upload files
		a.update(func() {
			modal := tview.NewModal().
				SetText("Would you like to upload files from the current directory as context?").
				AddButtons([]string{"Yes", "No"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					a.ui.pages.RemovePage("modal")

					if buttonLabel == "Yes" {
						a.input <- "yes"
					} else {
						a.input <- "no"
					}
				})

			a.ui.pages.AddPage("modal", modal, true, true)
			a.ui.app.SetFocus(modal)
		})

		var toUpload string
		select {
		case <-ctx.Done():
			return ctx.Err()
		case toUpload = <-a.input:
		}

		if toUpload == "yes" {
			a.state.filePathsInCwd, err = a.listFilePathsInCwd()
			if err != nil {
				return fmt.Errorf("a.listFilesInCwd(): %w", err)
			}

			if len(a.state.filePathsInCwd) > 0 {
				a.output("Files to upload:\n\n")

				for _, file := range a.state.filePathsInCwd {
					a.output(file + "\n")
				}

				mq := "are these files ok to upload?"

				a.output("\n\n" + mq)

				toUpload, err := a.readInput(ctx, mq+": ", "")
				if err != nil {
					return fmt.Errorf("a.readInput(): %w", err)
				}

				if toUpload == "yes" {
					a.update(func() {
						a.ui.output.SetText("")
					})

					a.output("Uploading files...\n")

					for _, filePath := range a.state.filePathsInCwd {
						resp, err := a.client.PostUpload(ctx, filePath)
						if err != nil {
							return fmt.Errorf("a.client.PostUpload(%s): %w", err)
						}

						a.output(fmt.Sprintf("\tuploaded: %s\n", filePath))

						attachmentUUIDs = append(attachmentUUIDs, resp.UUID)
					}

					a.output("\n\n...done uploading.\n\nawaiting question...\n\n")
				}
			}
		}

		question, err := a.readInput(ctx, "ask anything: ", "contemplating...")
		if err != nil {
			return fmt.Errorf("a.readInput(): %w", err)
		}

		convo, err := a.client.PostConversations(
			ctx,
			model,
			question,
			a.cache.LastReasoningEffort,
			attachmentUUIDs)
		if err != nil {
			return fmt.Errorf("client.PostConversations(): %w", err)
		}

		convoInit, err := a.client.GetConversationInit(ctx, convo.ConversationUUID)
		if err != nil {
			return fmt.Errorf("client.GetConversationInit(): %w", err)
		}

		var turnID string
		for _, turn := range convoInit.Turns {
			switch turn.Role {
			case "user":
			case "assistant":
				turnID = turn.UUID
			default:
				a.logger.Debugf("unknown turn role: %s", turn.Role)
			}
		}

		if turnID == "" {
			a.logger.Error("no assistant turns to stream")

			continue
		}

		eventChan, err := a.client.GetTurnStream(ctx, turnID)
		if err != nil {
			return fmt.Errorf("client.GetTurnStream(): %w", err)
		}

		err = a.HandleTurnStreamEvents(ctx, eventChan)
		if err != nil {
			return fmt.Errorf("a.HandleTurnStreamEvents(): %w", err)
		}

		// follow up loop
		for {
			a.output("\n\nthat's it folks! ask another question in the same thread or type \"new\"...\n\n")

			question, err := a.readInput(ctx, "ask anything: ", "contemplating...")
			if err != nil {
				return fmt.Errorf("a.readInput(): %w", err)
			}

			if question == "new" {
				break
			}

			branchResponse, err := a.client.PostBranchesRespond(
				ctx,
				convoInit.Branches[0].UUID,
				model,
				question)
			if err != nil {
				return fmt.Errorf("a.client.PostBranchesRespond(): %w", err)
			}

			eventChan, err := a.client.GetTurnStream(ctx, branchResponse.AssistantTurnUUID)
			if err != nil {
				return fmt.Errorf("client.GetTurnStream(): %w", err)
			}

			err = a.HandleTurnStreamEvents(ctx, eventChan)
			if err != nil {
				return fmt.Errorf("a.HandleTurnStreamEvents(): %w", err)
			}
		}
	}

	return nil
}

func (a *Agent) initUi() {
	flex := tview.NewFlex()
	flex.SetBorder(true)
	flex.SetTitle(" kagi-agent ")
	flex.SetDirection(tview.FlexRow)

	a.ui.header.SetBorder(true)
	a.ui.header.SetTitle(" thread ")
	a.ui.header.SetTitleAlign(tview.AlignLeft)
	flex.AddItem(a.ui.header, 4, 0, false)

	a.ui.output.ScrollToEnd()
	a.ui.output.SetBorder(true)
	a.ui.output.SetTitle(" output ")
	a.ui.output.SetTitleAlign(tview.AlignLeft)
	a.ui.pages.AddPage("output", a.ui.output, true, true)

	flex.AddItem(a.ui.pages, 0, 10, false)

	a.ui.input.SetBorder(true)
	a.ui.input.SetDoneFunc(a.inputDoneFunc)
	a.ui.input.SetTitleAlign(tview.AlignLeft)
	a.ui.input.SetTitle(" chat ")
	a.ui.input.SetDisabled(false)

	flex.AddItem(a.ui.input, 4, 0, true)

	a.ui.app.SetRoot(
		flex,
		// a.ui.pages,
		true, // fullScreen
	)
	// a.ui.app.EnableMouse(true)
	a.ui.app.EnablePaste(true)
}

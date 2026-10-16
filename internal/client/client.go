package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"charm.land/log/v2"

	"github.com/scrivy/kagi-agent/internal/client/models"
)

var (
	baseUri = "https://assistant.kagi.com"

	ErrUnauthorized = errors.New("http status 401 unauthorized")
)

type Client struct {
	logger *log.Logger
	c      *http.Client
}

func New(logger *log.Logger, sessionToken string) (*Client, error) {
	if sessionToken == "" {
		return nil, fmt.Errorf("session token cannot be empty")
	}

	u, err := url.Parse(baseUri)
	if err != nil {
		return nil, fmt.Errorf("url.Parse(): %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar.New(): %w", err)
	}

	jar.SetCookies(u, []*http.Cookie{
		&http.Cookie{
			Name:   "kagi_session",
			Value:  sessionToken,
			Secure: true,
		},
	})

	client := &Client{
		logger: logger,
		c: &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,
			},
			Jar: jar,
		},
	}

	return client, nil
}

func (c *Client) GetInit(ctx context.Context) (*models.GetInitResponse, error) {
	var init models.GetInitResponse

	body, status, err := c.do(ctx, http.MethodGet, "/api/init", nil, &init)
	if err != nil {
		return nil, fmt.Errorf("c.do(): %w", err)
	}

	switch status {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return &init, ErrUnauthorized
	default:
		return &init, fmt.Errorf("http status not ok: %d, body:\n%s", status, body)
	}

	return &init, nil
}

func (c *Client) PostConversations(
	ctx context.Context,
	model,
	question,
	reasoningEffort string,
	attachmentUUIDs []string) (*models.PostConversationsResponse, error) {

	input := models.PostConversationRequest{
		AttachmentUUIDs: attachmentUUIDs,
		Content:         question,
		EnableSearch:    true,
		LensID:          nil,
		Model:           model,
		Personalization: true,
		PlanRequested:   false,
		ReasoningEffort: reasoningEffort,
	}

	var convo models.PostConversationsResponse

	body, status, err := c.do(ctx, http.MethodPost, "/api/v2/conversations", input, &convo)
	if err != nil {
		return nil, fmt.Errorf("c.do(): %w", err)
	}

	if status != http.StatusOK {
		return &convo, fmt.Errorf("http status not ok: %d, body:\n%s", status, body)
	}

	return &convo, nil
}

func (c *Client) GetConversationInit(ctx context.Context, convoID string) (*models.GetConversationInit, error) {
	var init models.GetConversationInit

	body, status, err := c.do(
		ctx,
		http.MethodGet,
		"/api/v2/conversations/"+convoID+"/init",
		nil,
		&init)
	if err != nil {
		return nil, fmt.Errorf("c.do(): %w", err)
	}

	if status != http.StatusOK {
		return &init, fmt.Errorf("http status not ok: %d, body:\n%s", status, body)
	}

	return &init, nil
}

func (c *Client) PostBranchesRespond(
	ctx context.Context,
	branchUUID,
	model,
	question string) (*models.PostBranchesRespondResponse, error) {
	input := models.PostBranchesRespondRequest{
		Content:         question,
		PlanRequested:   false,
		Model:           model,
		EnableSearch:    true,
		LensID:          nil,
		Personalization: true,
	}

	var resp models.PostBranchesRespondResponse

	body, status, err := c.do(ctx, http.MethodPost, "/api/v2/branches/"+branchUUID+"/respond", input, &resp)
	if err != nil {
		return nil, fmt.Errorf("c.do(): %w", err)
	}

	if status != http.StatusOK {
		return &resp, fmt.Errorf("http status not ok: %d, body:\n%s", status, body)
	}

	return &resp, nil
}

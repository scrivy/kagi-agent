package models

// GET /api/v2/conversations/CONVERSATION_UUID/init

type GetConversationInit struct {
	Conversation Conversation `json:"conversation"`
	Branches     []Branch     `json:"branches"`
	ActiveBranch Branch       `json:"active_branch"`
	Turns        []Turn       `json:"turns"`
	TurnsHasMore bool         `json:"turns_has_more"`
	ContextUsage ContextUsage `json:"context_usage"`
}

type Branch struct {
	UUID                   string      `json:"uuid"`
	ConversationUUID       string      `json:"conversation_uuid"`
	BranchName             interface{} `json:"branch_name"`
	MessageCount           int         `json:"message_count"`
	HeadMessageUUID        string      `json:"head_message_uuid"`
	BranchPointParentUUID  interface{} `json:"branch_point_parent_uuid"`
	BranchChildMessageUUID string      `json:"branch_child_message_uuid"`
	IsDefault              bool        `json:"is_default"`
	UpdatedAt              any         `json:"updated_at"`
	ContentVersion         int         `json:"content_version"`
}

type Turn struct {
	UUID               string        `json:"uuid"`
	Role               string        `json:"role"`
	ConversationUUID   string        `json:"conversation_uuid"`
	BranchUUID         string        `json:"branch_uuid"`
	Content            string        `json:"content"`
	CreatedAt          any           `json:"created_at"`
	Attachments        []interface{} `json:"attachments"`
	Status             *string       `json:"status"`
	ErrorCode          interface{}   `json:"error_code"`
	CompletedAt        interface{}   `json:"completed_at"`
	Model              *string       `json:"model"`
	ModelVersion       interface{}   `json:"model_version"`
	InputTokens        interface{}   `json:"input_tokens"`
	OutputTokens       interface{}   `json:"output_tokens"`
	CostUsd            interface{}   `json:"cost_usd"`
	TokensPerSecond    interface{}   `json:"tokens_per_second"`
	DurationMs         interface{}   `json:"duration_ms"`
	Thinking           interface{}   `json:"thinking"`
	ThinkingHTML       interface{}   `json:"thinking_html"`
	HTMLContent        interface{}   `json:"html_content"`
	References         []interface{} `json:"references"`
	IsPinned           bool          `json:"is_pinned"`
	Annotation         interface{}   `json:"annotation"`
	ParentMessageUUID  *string       `json:"parent_message_uuid"`
	BranchName         interface{}   `json:"branch_name"`
	BranchLinks        interface{}   `json:"branch_links"`
	SiblingCount       int           `json:"sibling_count"`
	SiblingIndex       int           `json:"sibling_index"`
	SiblingUuids       []string      `json:"sibling_uuids"`
	SiblingBranchUuids []string      `json:"sibling_branch_uuids"`
	ProfileUUID        interface{}   `json:"profile_uuid"`
	ProfileRevisionID  interface{}   `json:"profile_revision_id"`
	ProfileName        interface{}   `json:"profile_name"`
	BuiltinProfileKey  interface{}   `json:"builtin_profile_key"`
	Events             []ConvoEvent  `json:"events"`
}

type ConvoEvent struct {
	ID           string       `json:"id"`
	Seq          int          `json:"seq"`
	TurnID       string       `json:"turn_id"`
	Timestamp    any          `json:"timestamp"`
	Source       string       `json:"source"`
	Type         string       `json:"type"`
	NodeID       *string      `json:"node_id"`
	ParentNodeID interface{}  `json:"parent_node_id"`
	Payload      ConvoPayload `json:"payload"`
}

type ConvoPayload struct {
	Tier       *string `json:"tier"`
	BranchUUID *string `json:"branch_uuid"`
	RootNodeID *string `json:"root_node_id"`
	Kind       *string `json:"kind"`
	Status     *string `json:"status"`
}

type ContextUsage struct {
	ContextWindow   int `json:"context_window"`
	SystemTokens    int `json:"system_tokens"`
	UserTokens      int `json:"user_tokens"`
	AssistantTokens int `json:"assistant_tokens"`
	TotalUsed       int `json:"total_used"`
	Available       int `json:"available"`
	UsedPercent     int `json:"used_percent"`
}

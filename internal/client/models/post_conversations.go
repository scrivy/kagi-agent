package models

// POST /api/v2/conversations
type PostConversationRequest struct {
	AttachmentUUIDs []string `json:"attachment_uuids,omitempty"`
	Content         string   `json:"content"`       // question
	EnableSearch    bool     `json:"enable_search"` // true
	LensID          *string  `json:"lens_id"`
	Model           string   `json:"model"`
	Personalization bool     `json:"personalization"`            // true
	PlanRequested   bool     `json:"plan_requested"`             // false
	ReasoningEffort string   `json:"reasoning_effort,omitempty"` // for deepseek: none, low, high (default), max
}

type PostConversationsResponse struct {
	ConversationUUID  string       `json:"conversation_uuid"`
	BranchUUID        string       `json:"branch_uuid"`
	AssistantTurnUUID string       `json:"assistant_turn_uuid"`
	UserTurnUUID      string       `json:"user_turn_uuid"`
	Conversation      Conversation `json:"conversation"`
}

type Conversation struct {
	UUID              string  `json:"uuid"`
	Title             string  `json:"title"`
	IsSaved           bool    `json:"is_saved"`
	TemporaryDuration string  `json:"temporary_duration"`
	IsShared          bool    `json:"is_shared"`
	IsPinned          bool    `json:"is_pinned"`
	ModelName         string  `json:"model_name"`
	Icon              *string `json:"icon"`
	FolderUUID        *string `json:"folder_uuid"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	DeletedAt         *string `json:"deleted_at"`
	TotalTokens       int     `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	PinnedCount       int     `json:"pinned_count"`
	MessageCount      int     `json:"message_count"`
	AnnotationCount   int     `json:"annotation_count"`
	HasBranches       bool    `json:"has_branches"`
	BranchCount       int     `json:"branch_count"`
	IsOwner           bool    `json:"is_owner"`
}

package models

// POST /api/v2/branches/BRANCH_UUID/respond

type PostBranchesRespondRequest struct {
	Content         string  `json:"content"`
	PlanRequested   bool    `json:"plan_requested"` // false
	Model           string  `json:"model"`
	EnableSearch    bool    `json:"enable_search"`   // true
	LensID          *string `json:"lens_id"`         // nil
	Personalization bool    `json:"personalization"` // true
}

type PostBranchesRespondResponse struct {
	AssistantTurnUUID string `json:"assistant_turn_uuid"`
	ConversationUUID  string `json:"conversation_uuid"`
	BranchUUID        string `json:"branch_uuid"`
	UserTurnUUID      string `json:"user_turn_uuid"`
}

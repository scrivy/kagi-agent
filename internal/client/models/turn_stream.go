package models

// GET /api/v2/turns/TURN_UUID/init

type TurnStreamEvent struct {
	Event Event `json:"event"`
}

type Event struct {
	ID               string           `json:"id"`
	Seq              int              `json:"seq"`
	TurnID           string           `json:"turn_id"`
	Timestamp        string           `json:"timestamp"`
	Source           string           `json:"source"`
	Type             string           `json:"type"`
	NodeID           string           `json:"node_id"`
	ParentNodeID     string           `json:"parent_node_id"`
	Payload          TurnEventPayload `json:"payload"`
	ConversationUUID string           `json:"conversation_uuid"`
	BranchUUID       string           `json:"branch_uuid"`
}

type TurnEventPayload struct {
	Status  string `json:"status"`
	Content string `json:"content"`
}

package models

// GET /api/init
type GetInitResponse struct {
	Conversations       Conversations       `json:"conversations"`
	Counts              Counts              `json:"counts"`
	Folders             []Folder            `json:"folders"`
	FolderCustomization FolderCustomization `json:"folder_customization"`
	Models              Models              `json:"models"`
	Lenses              []Lens              `json:"lenses"`
	Settings            Settings            `json:"settings"`
	CustomAssistants    []interface{}       `json:"custom_assistants"`
	ActiveConversation  interface{}         `json:"active_conversation"`
	ActiveError         interface{}         `json:"active_error"`
	Billing             Billing             `json:"billing"`
	HasUnlimitedAccess  bool                `json:"has_unlimited_access"`
	IsStaff             bool                `json:"is_staff"`
	FeatureFlags        []string            `json:"feature_flags"`
	Features            Features            `json:"features"`
}

type Conversations struct {
	Items   []Item `json:"items"`
	HasMore bool   `json:"has_more"`
}

type Item struct {
	UUID              string      `json:"uuid"`
	Title             string      `json:"title"`
	IsSaved           bool        `json:"is_saved"`
	TemporaryDuration string      `json:"temporary_duration"`
	IsShared          bool        `json:"is_shared"`
	IsPinned          bool        `json:"is_pinned"`
	ModelName         string      `json:"model_name"`
	Icon              interface{} `json:"icon"`
	FolderUUID        interface{} `json:"folder_uuid"`
	CreatedAt         any         `json:"created_at"`
	UpdatedAt         any         `json:"updated_at"`
	DeletedAt         interface{} `json:"deleted_at"`
	TotalTokens       int         `json:"total_tokens"`
	TotalCost         float64     `json:"total_cost"`
	PinnedCount       int         `json:"pinned_count"`
	MessageCount      int         `json:"message_count"`
	AnnotationCount   int         `json:"annotation_count"`
	HasBranches       bool        `json:"has_branches"`
	BranchCount       int         `json:"branch_count"`
	IsOwner           bool        `json:"is_owner"`
}

type Counts struct {
	All        int `json:"all"`
	Unassigned int `json:"unassigned"`
	Deleted    int `json:"deleted"`
}

type Folder struct {
	UUID              string      `json:"uuid"`
	Name              string      `json:"name"`
	Color             interface{} `json:"color"`
	Icon              interface{} `json:"icon"`
	Position          int         `json:"position"`
	CreatedAt         any         `json:"created_at"`
	UpdatedAt         any         `json:"updated_at"`
	ConversationCount int         `json:"conversation_count"`
}

type FolderCustomization struct {
	AvailableColors []string `json:"available_colors"`
	AvailableIcons  []string `json:"available_icons"`
}

type Models struct {
	Models     []Model     `json:"models"`
	Assistants []Assistant `json:"assistants"`
	Default    string      `json:"default"`
	Sections   []Section   `json:"sections"`
}

type Model struct {
	ID               string     `json:"id"`
	Provider         string     `json:"provider"`
	ProviderLabel    string     `json:"provider_label"`
	DisplayName      string     `json:"display_name"`
	ContextWindow    int        `json:"context_window"`
	SupportsThinking bool       `json:"supports_thinking"`
	Parameters       Parameters `json:"parameters"`
	Capabilities     []string   `json:"capabilities"`
	Supported        bool       `json:"supported"`
	Recommended      bool       `json:"recommended"`
	Scorecard        Scorecard  `json:"scorecard"`
	InternetAccess   bool       `json:"internet_access"`
	RequiresSearch   bool       `json:"requires_search"`
	AccessLevel      string     `json:"access_level"`
	Deprecated       bool       `json:"deprecated"`
	Retired          bool       `json:"retired"`
	SuccessorModelID *string    `json:"successor_model_id"`
}

type Parameters struct {
	ReasoningEffort *ReasoningEffort `json:"reasoning_effort"`
}

type ReasoningEffort struct {
	Options []string `json:"options"`
	Default *string  `json:"default"`
}

type Scorecard struct {
	Cost          int            `json:"cost"`
	Speed         int            `json:"speed"`
	Accuracy      int            `json:"accuracy"`
	Privacy       int            `json:"privacy"`
	ReleaseDate   any            `json:"release_date"`
	QualitySource *QualitySource `json:"quality_source"`
}

type QualitySource struct {
	Citation  string `json:"citation"`
	SourceURL string `json:"source_url"`
	AsOf      any    `json:"as_of"`
}

type Assistant struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	UnderlyingModel    string      `json:"underlying_model"`
	Description        string      `json:"description"`
	InternetAccess     bool        `json:"internet_access"`
	Personalizations   bool        `json:"personalizations"`
	LensID             interface{} `json:"lens_id"`
	Retired            bool        `json:"retired"`
	Deprecated         bool        `json:"deprecated"`
	SuccessorModelName interface{} `json:"successor_model_name"`
}

type Section struct {
	Key      string   `json:"key"`
	Label    string   `json:"label"`
	ModelIds []string `json:"model_ids"`
}

type Lens struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Settings struct {
	ThreadSavingMode         string      `json:"thread_saving_mode"`
	ThreadRetentionPolicy    string      `json:"thread_retention_policy"`
	TemporaryDuration        string      `json:"temporary_duration"`
	CustomInstructions       interface{} `json:"custom_instructions"`
	DefaultSelectionKind     string      `json:"default_selection_kind"`
	DefaultCustomProfileUUID interface{} `json:"default_custom_profile_uuid"`
	DefaultBuiltinProfileKey interface{} `json:"default_builtin_profile_key"`
	DefaultLlmID             interface{} `json:"default_llm_id"`
	WorkspaceArea            interface{} `json:"workspace_area"`
	FontSize                 interface{} `json:"font_size"`
	LightTheme               interface{} `json:"light_theme"`
	DarkTheme                interface{} `json:"dark_theme"`
	KeyboardShortcuts        interface{} `json:"keyboard_shortcuts"`
	SubmitKey                interface{} `json:"submit_key"`
	InterfaceLanguage        string      `json:"interface_language"`
	UpsellDismissed          bool        `json:"upsell_dismissed"`
	FolderMoveSave           bool        `json:"folder_move_save"`
}

type Billing struct {
	Limit       string      `json:"limit"`
	Reason      interface{} `json:"reason"`
	CanProceed  bool        `json:"can_proceed"`
	LimitStatus string      `json:"limit_status"`
	Usage       Usage       `json:"usage"`
	Ack         Ack         `json:"ack"`
}

type Usage struct {
	InteractionCount         int         `json:"interaction_count"`
	InteractionLimit         interface{} `json:"interaction_limit"`
	InteractionRemaining     interface{} `json:"interaction_remaining"`
	IncludedCostUsd          float64     `json:"included_cost_usd"`
	IncludedCostLimitUsd     float64     `json:"included_cost_limit_usd"`
	IncludedCostSoftLimitUsd float64     `json:"included_cost_soft_limit_usd"`
	PaygCostUsd              interface{} `json:"payg_cost_usd"`
	PaygHardLimitUsd         interface{} `json:"payg_hard_limit_usd"`
}

type Ack struct {
	Soft bool `json:"soft"`
	Hard bool `json:"hard"`
}

type Features struct {
	AgentV2          bool `json:"agent_v2"`
	NewAstExperience bool `json:"new_ast_experience"`
}

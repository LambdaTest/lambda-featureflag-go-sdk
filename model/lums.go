package model

type OrgId int64

type OrgMap map[OrgId]OrgId

// ErrorResponse represents the error response from the LUMS API
type ErrorResponse struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// SuccessResponse represents the success response from the LUMS API
type SuccessResponse struct {
	Type string `json:"type"`
	Data OrgMap `json:"data"`
}

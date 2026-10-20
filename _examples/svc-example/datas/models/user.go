package models

// UserSearchReq provides parameters for complex queries.
type UserSearchReq struct {
	Page     int32
	Size     int32
	Sort     string
	Keywords string
}

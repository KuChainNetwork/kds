package types

import "encoding/json"

// PageRequest 页请求
type PageRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Status   int `json:"status"`
}

// Offset 偏移量
func (object *PageRequest) Offset() int {
	if 0 >= object.Page {
		return 0
	}
	return (object.Page - 1) * object.PageSize
}

// Limit 限制
func (object *PageRequest) Limit() int {
	return object.PageSize
}

// String 字符串描述
func (object *PageRequest) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}

// PageResponse 页响应
type PageResponse struct {
	Total int         `json:"total"`
	List  interface{} `json:"list"`
}

// String 字符串描述
func (object *PageResponse) String() string {
	raw, _ := json.Marshal(object)
	return string(raw)
}

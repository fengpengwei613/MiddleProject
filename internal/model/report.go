package model

import (
	_ "github.com/go-sql-driver/mysql"
)
//举报结构请求体
type ReportRequest struct {
    LogID string `json:"logid,omitempty"`
    ComID string `json:"commentid,omitempty"`
    ReplyID string `json:"replyid,omitempty"`
	Reason string `json:"reason"`
    Type  string `json:"type"`
}
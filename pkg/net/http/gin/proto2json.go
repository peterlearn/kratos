package gin

import (
	"github.com/peterlearn/kratos/v1/pkg/ecode"
)

type JSON struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	TTL     int         `json:"ttl"`
	Data    interface{} `json:"data,omitempty"`
}

// 默认TOJSON
func TOJSON(resp interface{}, err error) (data JSON) {
	bcode := ecode.Cause(err)

	return JSON{
		Code:    bcode.Code(),
		Message: bcode.Message(),
		TTL:     1,
		Data:    resp,
	}
}

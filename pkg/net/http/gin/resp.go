package gin

import (
	"context"
	"github.com/peterlearn/kratos/v1/pkg/ecode"
)

const (
	RespJson     = 0 //返回json 默认返回json
	RespProtobuf = 1 //返回protobuf二进制
)

var (
	ToResponse func(context.Context, interface{}, error) (interface{}, int)
)

func init() {
	ToResponse = defaultResponse
}

// 默认返回
func defaultResponse(ctx context.Context, resp interface{}, err error) (interface{}, int) {
	bcode := ecode.Cause(err)

	return JSON{
		Code:    bcode.Code(),
		Message: bcode.Message(),
		TTL:     1,
		Data:    resp,
	}, RespJson
}

// 自定义返回
func SetResponse(fc func(ctx context.Context, resp interface{}, err error) (interface{}, int)) {
	ToResponse = fc
}

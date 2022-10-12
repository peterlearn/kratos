# 参数验证模块

# 介绍

在接口调用过程中，有一块内容非常重要，却很容易被我们忽略：**参数验证**，忘了之后常常会给我们造成大量的处理错误问题，甚至直接造成应用崩溃。

**如果需要我们自己验证**

显然，肯定是有的，而且跟我们又爱又恨的 reflect 有关系，因为对于强类型语言来说，如果没有元编程的帮助，对于数据的验证会变得非常难受：

```go
if strings.Trim(mobile, " ") == "" {
    return nil, ErrEmptyParams
}
mobileRegex := regexp.MustCompile(`^1\d{10}$`)
if !mobileRegex.MatchString(mobile) {  
    return nil, ErrInvalidMobile
}
```

上面这类似的代码会充斥在你的业务逻辑处理当中，非常不优雅，即使我们可以把常见的验证封装，也不能避免我们这样的情况，充其量只能缩短验证的代码行数而已。

回过头来，当我们有了 reflect 的帮助，验证就会变得非常简单，我们只要在 `Struct Tag` 中，声明字段的验证要求，剩下的，就交给框架或者中间件去统一处理即可，非常简单而又优雅，并且容易维护。

Kratos框架引用了golang的validator库：[go-playground/validator](https://github.com/go-playground/validator)，在http和grpc模块的中间件里做统一处理，使用起来非常简单，只需在定义protobuf协议时对参数设置验证规则即可。

## Example

**1、定义protobuf**

```protobuf
//斗地主请求参数
message DdzParamsReq
{
    string uid = 1 [(gogoproto.moretags) = 'validate:"required"'];
    string tid = 2 [(gogoproto.moretags) = 'validate:"required"'];
    string ids = 3 [(gogoproto.moretags) = 'validate:"required"'];
    string order = 4 [(gogoproto.moretags) = 'validate:"required"'];
    string role = 5 [(gogoproto.moretags) = 'validate:"required"'];
    string record = 6 [(gogoproto.moretags) = 'validate:"required"'];
    string hands = 7 [(gogoproto.moretags) = 'validate:"required"'];
    string up_hands = 8 [(gogoproto.moretags) = 'validate:"required"'];
    string down_hands = 9 [(gogoproto.moretags) = 'validate:"required"'];
    string hold_up = 10;
    string is_special = 11;
    int64 timestamp = 12;
}
```

**2、使用kratos tool工具生成接口代码**

`kratos tool protoc api.proto`

```go
//斗地主请求参数
type DdzParamsReq struct {
	Uid                  string   `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty" validate:"required"`
	Tid                  string   `protobuf:"bytes,2,opt,name=tid,proto3" json:"tid,omitempty" validate:"required"`
	Ids                  string   `protobuf:"bytes,3,opt,name=ids,proto3" json:"ids,omitempty" validate:"required"`
	Order                string   `protobuf:"bytes,4,opt,name=order,proto3" json:"order,omitempty" validate:"required"`
	Role                 string   `protobuf:"bytes,5,opt,name=role,proto3" json:"role,omitempty" validate:"required"`
	Record               string   `protobuf:"bytes,6,opt,name=record,proto3" json:"record,omitempty" validate:"required"`
	Hands                string   `protobuf:"bytes,7,opt,name=hands,proto3" json:"hands,omitempty" validate:"required"`
	UpHands              string   `protobuf:"bytes,8,opt,name=up_hands,json=upHands,proto3" json:"up_hands,omitempty" validate:"required"`
	DownHands            string   `protobuf:"bytes,9,opt,name=down_hands,json=downHands,proto3" json:"down_hands,omitempty" validate:"required"`
	HoldUp               string   `protobuf:"bytes,10,opt,name=hold_up,json=holdUp,proto3" json:"hold_up,omitempty"`
	IsSpecial            string   `protobuf:"bytes,11,opt,name=is_special,json=isSpecial,proto3" json:"is_special,omitempty"`
	Timestamp            int64    `protobuf:"varint,12,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
```

**3、使用kratos脚手架工具生成项目**

`kratos new [serverName] --proto`

**4、编译项目，启动服务**

在调用接口时，如参数验证不通过，则服务会返回相应的错误提示

```json
{
    "code": -400,
    "message": "Key: 'DdzParamsReq.Uid' Error:Field validation for 'Uid' failed on the 'required' tag",
    "ttl": 1
}
```

## validate验证规则介绍

**字段验证**

常见的邮箱、非空、最大最小、长度限制等验证

```go
type Test struct {
  Email string `validate:"email"`
  Size  int    `validate:"max=10,min=1"`
}
```

**跨字段以及跨Struct验证**

对于字段之间，甚至跨 Struct 之间的字段验证，它都可以做到，主要有：

- `eqfield=Field`: 必须等于 Field 的值；
- `nefield=Field`: 必须不等于 Field 的值；
- `gtfield=Field`: 必须大于 Field 的值；
- `gtefield=Field`: 必须大于等于 Field 的值；
- `ltfield=Field`: 必须小于 Field 的值；
- `ltefield=Field`: 必须小于等于 Field 的值；
- `eqcsfield=Other.Field`: 必须等于 struct Other 中 Field 的值；
- `necsfield=Other.Field`: 必须不等于 struct Other 中 Field 的值；
- `gtcsfield=Other.Field`: 必须大于 struct Other 中 Field 的值；
- `gtecsfield=Other.Field`: 必须大于等于 struct Other 中 Field 的值；
- `ltcsfield=Other.Field`: 必须小于 struct Other 中 Field 的值；
- `ltecsfield=Other.Field`: 必须小于等于 struct Other 中 Field 的值；

```go
type Test struct {
	StartAt time.Time `validate:"required"`
	EndAt   time.Time `validate:"required,gtfield=StartAt"`
}
```

更多规则说明参考[文档](https://godoc.org/gopkg.in/go-playground/validator.v9)。
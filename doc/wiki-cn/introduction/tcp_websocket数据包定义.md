## Tcp/WebSocket数据包定义

### Body（protobuf）

| PlayerId | Ops   | Data  |
| -------- | ----- | ----- |
| int64    | int32 | bytes |

### Request

| PackLen | Place（占位，无用） | Type（Request：1） | Seq   | Body         |
| ------- | ------------------- | ------------------ | ----- | ------------ |
| 2byte   | 1byte               | 1byte              | 2byte | 最大4096byte |

### Response

| PackLen | Place | Type（Response：2） | Seq   | Code（错误码） | Body         |
| ------- | ----- | ------------------- | ----- | -------------- | ------------ |
| 2byte   | 1byte | 1byte               | 2byte | 2byte          | 最大4096byte |

### Push

| PackLen | Place | Type（Push：0） | Body         |
| ------- | ----- | --------------- | ------------ |
| 2byte   | 1byte | 1byte           | 最大4096byte |

### HeartBeat

| PackLen | Place | Type（Ping：3，Pong：4） | Seq   |
| ------- | ----- | ------------------------ | ----- |
| 2byte   | 1byte | 1byte                    | 1byte |

### 




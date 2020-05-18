# tt_tac
---
## 项目介绍
1. 从以太坊的token跨链转移到tt链的相同token，手续费暂定1 token
2. 从tt链的token转移到以太坊上的相同token，手续费暂定3 token
3. 目前系统还不支持两种token的小数位不相等的情况的跨链转账(目前觉得没有必要优化，也没有时间优化)
4. 由于以太坊的网络比较拥挤，所以每个跨链转账订单完成时间有长有短


---    

## tac 接口文档
---
#### 创建一个跨链转账订单
###### 请求url
- POST `/tac/apply_order`
###### 请求参数
```$json
{
	"fromAddr":"0x67Adf250F70F6100d346cF8FE3af6DC7A2C23213",
	"recipientAddr":"0x7AC954Ed6c2d96d48BBad405aa1579C828409f59",
	"amount":"5555000000000",
	"orderType":2
}
```
###### 参数说明
1. `fromAddress`: pala代币转出地址
2. `recipientAddr`: pala代币接收地址
3. `amount`: 跨链转账数量，后面必须加上8个0
4. `orderType`: 跨链转账的类型，为1表示从以太坊的pala转到tt链的pala, 为2表示从tt链的pala转到以太坊上的pala
###### 返回示例
```$xslt
// 成功返回示例，会返回订单号给前端
{
    "status": 200,
    "data": {
        "orderId": 5
    },
    "msg": "success",
    "error": ""
}
```
---
#### 通过订单号查询订单详情
###### 请求url
- GET `/tac/order/:id`
###### 请求参数
`id`: 订单id
###### 返回示例
```$xslt
// 接口调用成功
{
    "status": 200,
    "data": {
        "from_addr": "0x67adf250f70f6100d346cf8fe3af6dc7a2c23213",
        "recipient_addr": "0x7ac954ed6c2d96d48bbad405aa1579c828409f59",
        "amount": "5555000000000",
        "order_type": 1,
        "state": 0
    },
    "msg": "success",
    "error": ""
}
```
###### 字段说明
1. `from_addr`: 跨链pala代币转出地址
2. `recipient_addr`: pala代币接收地址
3. `amount`: 转账金额
4. `order_type`: 订单类型，为1表示从以太坊的pala转到tt链的pala, 为2表示从tt链的pala转到以太坊上的pala
5. `state`: 订单状态，为0表示订单进行中，为1表示订单完成，为2表示订单失败，为3表示订单超时
---
#### 创建user
###### 请求url
- POST `/tac/create_user`
###### 请求参数
```$xslt
{
	"password":"123456"
}
```
###### 参数说明
1. password: 用户支付时的密码，最小6位，最大12位
###### 返回示例
```$xslt
// 返回地址address,需要前端保存到localstorage中，要求用户需要备份地址。
{
    "status": 200,
    "data": {
        "address": "0xb86ebA9f29Fcc6cA8dE202889111dC1c6BEdDf16"
    },
    "msg": "success",
    "error": ""
}
```
----
#### 通过私钥导入user
###### 请求url
- POST /tac/lead_user
###### 请求参数
```$xslt
{
	"private":"90909E90903DCCF0A03D9D1BE998E161532A264A959C8989158B6C9ACA92H33C",
	"password":"12345678"
}
```
###### 参数说明
1. private：导入的私钥
2. password：账户支付密码设置
###### 返回示例
```$xslt
// 返回导入private对应的address，前端需要把address 存入localstorage
{
    "status": 200,
    "data": {
        "address": "0x67Adf250F70F6100d346cF8FE3af6DC7A2C99999"
    },
    "msg": "success",
    "error": ""
}
```
----


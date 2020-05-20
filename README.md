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

#### 发送跨链转账交易
###### 请求url
- POST `/tac/send_tac_tx`
###### 请求参数
```$xslt
{
	"address": "0x59375A522876aB96B0ed2953D0D3b92674701Cc2",
	"password":"123456",
	"amount":"911000000",
	"order_type":2
}
```
###### 参数说明
1. address: 钱包地址
2. password: 钱包支付密钥
3. amount: 跨链金额
4. order_type: 跨链类型。1表示从eth_pala转到tt_pala，2则相反
###### 返回示例
```$xslt
// 返回交易hash
{
    "status": 200,
    "data": {
        "tx_hash": "0x9687dc46485f0be3791d8054c39a8d3b9f10ac108f7d8157494ae0391e671081"
    },
    "msg": "success",
    "error": ""
}
```

#### 创建wallet
###### 请求url
- POST `/tac/create_wallet`
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
#### 通过私钥导入wallet
###### 请求url
- POST /tac/lead_wallet
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
#### 导出私钥
###### 请求url
- POST /tac/export_private
###### 请求参数
```$xslt
{
	"address":"0xb86ebA9f29Fcc6cA8dE202889111dC1c6BEdDf16",
	"password":"123456"
	
}
```
###### 参数说明
1. address: 钱包地址
2. password: 支付密码
###### 返回示例
```$xslt
// 成功返回示例
{
    "status": 200,
    "data": {
        "private": "F234120DE07D7F5CE27EAA1D7B954F55BDC49E6C3B2B19FB78C5000A191CEE4F"
    },
    "msg": "success",
    "error": ""
}

// password失败返回示例
{
    "status": 40006,
    "data": null,
    "msg": "导出私钥失败",
    "error": "输入的密码有误"
}
// address有误的返回示例
{
    "status": 40006,
    "data": null,
    "msg": "导出私钥失败",
    "error": "record not found"
}
```
---
#### 申请闪兑接口
###### 请求url
- POST /tac/exchange/eth_usdt_pala
###### 请求参数
```$xslt
{
	"operate_address":"0x7AC954Ed6c2d96d48BBad405aa1579C828409f59",
	"password":"123456",
	"from_token_amount":"5000000000000000",
	"to_token_amount":"1000000000000000000"
	
}
```
###### 参数说明
1. operate_address: 需要闪兑的地址，也是用户的钱包地址
2. password: 钱包的支付密码
3. from_token_amount: usdt的兑换amount
4. to_token_amount: 兑换成pala的amount
###### 返回示例
```$xslt
{
    "status": 200,
    "data": {
        "tx_hash": "0x6cd33023cedffd006a70d5c2225006a51415a90398673a979ad5dbf3542bc2b5"
    },
    "msg": "success",
    "error": ""
}
```
----
#### 拉取地址的tt和eth上的主网币余额
###### 请求url
- POST /tac/get_balance
###### 请求参数
```$xslt
{
	"address":"0x67Adf250F70F6100d346cF8FE3af6DC7A2C23213"
}
```
###### 返回示例
```$xslt
{
    "status": 200,
    "data": {
        "tt_balance": "747870339000000000",
        "eth_balance": "19000000000000",
        "decimal": 18
    },
    "msg": "success",
    "error": ""
}
```

----
#### 拉取地址的tt和eth上token的余额
###### 请求url
- POST /tac/get_token_balance
###### 请求参数
```$xslt
{
	"address":"0x7AC954Ed6c2d96d48BBad405aa1579C828409f59"
}
```
###### 返回示例
```$xslt
{
    "status": 200,
    "data": {
        "tt_pala_balance": "100000000",
        "eth_pala_balance": "101000008999990000000",
        "eth_usdt_balance": "35000000000000000",
        "usdt_decimal": 6,
        "pala_decimal": 8
    },
    "msg": "success",
    "error": ""
}
```
###### 参数说明
1. tt_pala_balance: tt链上pala的余额
2. eth_pala_balance: eth链上pala的余额
3. eth_usdt_balance: eth链上usdt的余额
4. usdt_decimal: usdt小数位数
5. pala_decimal: pala小数位数
---
#### 获取eth链上的pala的实时价格
###### 请求url
- GET /tac/get_eth_pala_price
###### 返回示例
```$xslt
{
    "status": 200,
    "data": {
        "pair": "PALA_USDT",
        "trade_price": "6.31400000"
    },
    "msg": "success",
    "error": ""
}
```
---
#### 获取eth的实时价格
###### 请求url
- GET /tac/get_eth_price
###### 返回示例
```$xslt
{
    "status": 200,
    "data": {
        "pair": "ETH_USDT",
        "trade_price": "213.69"
    },
    "msg": "success",
    "error": ""
}
```
---

# tt_tac
---
## 项目介绍
#### 钱包
###### 功能
1. 创建钱包，设置支付密码
2. 通过私钥导入钱包，设置支付密码
3. 导出私钥，支付密码验证
4. tt链和以太坊共用同一个账户钱包体系

#### 跨链转账
###### 功能
1. 目前实现了tt链和以太坊上的pala erc20代币相互之间的转移
2. 用户在钱包页面操作跨链，前端调取后端两个接口，第一个是生成一笔跨链转账订单的接口，返回成功状态之后再调取发送跨链转账交易的接口
3. 前端通过转账订单返回的订单号来拉取跨链转账的转账详情展示给用户
4. 跨链转账需要扣除一定的pala作为转账的交易手续费,所以接收的pala会少于发起的pala数量
5. 用户操作跨链转账的账户必须要有足够的eth或者tt作为发起交易的gas fee

#### usdt闪兑pala
###### 功能
1. 用户可以在该钱包中实现以太坊上的usdt闪兑成以太坊上的pala代币
2. 前端调取申请闪兑接口即可开启闪兑流程
3. 前端通过返回的闪兑的订单号来拉取闪兑的进度展示给用户
4. 前端拉取后端的token价格的接口，计算出闪兑的兑换比例
5. 用户执行闪兑操作，账户中必须有足够的eth用于发送闪兑交易的gas fee

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
#### 修改支付密码
###### 请求url
- POST /tac/modify_password
###### 请求参数
```
{
	"address":"0x7AC954Ed6c2d96d48BBad405aa1579C828409f59",
	"old_password":"12345678",
	"new_password":"123456789"
}
```
###### 参数说明
1. address: 修改密码地址
2. old_password: 旧密码
3. new_password: 新密码
###### 返回示例
```$xslt
// 成功情况
{
    "status": 200,
    "data": null,
    "msg": "success",
    "error": ""
}
// 失败情况1. 旧密码不正确
{
    "status": 40014,
    "data": null,
    "msg": "modify password error",
    "error": "旧的密码验证不通过"
}
// 失败情况2. address不存在
{
    "status": 40014,
    "data": null,
    "msg": "modify password error",
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
	"to_token_amount":"1000000000000000000",
	"trade_price":"6.771"
	
}
```
###### 参数说明
1. operate_address: 需要闪兑的地址，也是用户的钱包地址
2. password: 钱包的支付密码
3. from_token_amount: usdt的兑换amount
4. to_token_amount: 兑换成pala的amount
5. trade_price: 闪兑的价格
###### 返回示例
```$xslt
{
    "status": 200,
    "data": {
        "orderId": 5
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
        "tt_balance": "7478703.390000",
        "eth_balance": "1900.000000",
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
        "tt_pala_balance": "1000.00000",
        "eth_pala_balance": "101000.00899",
        "eth_usdt_balance": "35000.000000",
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
#### 分页拉取地址下面闪兑订单列表
###### 请求url
- POST /tac/exchange/get_flash_orders
###### 请求参数
```$xslt
{
	"start_index":0,
	"limit":5,
	"address":"0x7AC954Ed6c2d96d48BBad405aa1579C828409f59"
}
```
###### 参数说明
1. start_index: 分页的起始位置
2. limit: 每次拉取的数量
3. address: 地址
###### 返回示例
```$xslt
{
    "status": 200,
    "data": [
        {
            "created_at": 1590104685,
            "amount": "0.777770",
            "state": 1
        },
        {
            "created_at": 1590104685,
            "amount": "10000.000000",
            "state": 1
        },
        {
            "created_at": 1590104685,
            "amount": "100.000000",
            "state": 0
        }
    ],
    "msg": "success",
    "error": ""
}
```
###### 返回字段说明
1. created_at: 订单创建时间
2. amount: 闪兑的数量
3. state: 订单状态，0. pending，1. success 2. failed
---


#### 分页拉取跨链转账的订单记录
###### 请求url
- POST /tac/get_tac_orders
###### 请求参数
```$xslt
{
	"order_type":1,
	"address":"0x67adf250f70f6100d346cf8fe3af6dc7a2c23213",
	"start_index":1,
	"limit":5
}
```
###### 参数说明
1. order_type: 订单类型，orderType == 1 表示拉取以太坊跨链转账到tt链的订单，为2则相反
2. address: 对应地址
3. start_index：分页起始位置
4. limit:每一次拉取的条数
###### 返回示例
```$xslt
{
    "status": 200,
    "data": [
        {
            "created_at": 1590104685,
            "amount": "0.000777",
            "state": 1
        },
        {
            "created_at": 1590104685,
            "amount": "0.000666",
            "state": 0
        }
    ],
    "msg": "success",
    "error": ""
}
```
###### 返回字段说明
1. created_at: 订单创建时间
2. amount: 跨链pala的数量
3. state: 订单状态，0. pending，1. success 2. failed
----
#### 获取用户地址下的eth主网上pala接收记录
###### 请求url
- POST /tac/get_eth_pala_receive
###### 请求参数
```$xslt
{
	"address":"0x9d7bc48d1c7a42b5fa9e070b4e301d2445bea926"
}
```
###### 返回示例
```$xslt
{
    "status": 200,
    "data": [
        {
            "from": "0x65b1c87aa01c82c1d15adcda7e21f3187594b2c9",
            "to": "0x9d7bc48d1c7a42b5fa9e070b4e301d2445bea926",
            "amount": "478388000000",
            "time_at": 1590104685
        },
        {
             "from": "0x65b1c87aa01c82c1d15adcda7e21f3187594b2c9",
             "to": "0x9d7bc48d1c7a42b5fa9e070b4e301d2445bea926",
             "amount": "88888888888",
             "time_at": 1590104681
         },
    ],
    "msg": "success",
    "error": ""
}

// 记录为空的情况
{
    "status": 200,
    "data": [],
    "msg": "success",
    "error": ""
}
```
----
#### 获取发送一笔以太坊token转账交易或者tt的token转账交易需要的gas fee
###### 请求url
- POST /tac/get_gas_fee
###### 请求参数
```$xslt
{
	"chain_tag": 17
}
```
###### 参数说明
1. chain_tag: 链的标识，17代表以太坊的链，77代表thundercore链
###### 返回示例
```$xslt
{
    "status": 200,
    "data": {
        "gas_fee": "0.000120"
    },
    "msg": "success",
    "error": ""
}
```
---


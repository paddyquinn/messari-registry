# Prerequisites
go - https://golang.org/doc/install
dep - https://github.com/golang/dep

# Compilation
```
dep ensure
go build main.go
```
Note: https://github.com/mattn/go-sqlite3 is used as the SQLite driver. It requires cgo to build. cgo should be enabled
by default but try the following if you run into build issues with cgo:
```
export CC=gcc
export CGO_ENABLED=1
```

# Running the server
`./main`

# Running the unit tests
`go test ./...`

# Register Examples
```
$ curl -X POST localhost:8080/register -d '{'
{
  "error":"unexpected EOF"
}
$ curl -X POST localhost:8080/register -d '{}'
{
  "error":"team cannot be null"
}
$ curl -X POST localhost:8080/register -d '{"team": [], "foundedDate": "01/03/2009"}'
{
  "error":"date must be an ISO-8601 date in the past"
}
$ curl -X POST localhost:8080/register -d '{"team": [], "foundedDate": "2140-01-03"}'
{
  "error":"date must be an ISO-8601 date in the past"
}
$ curl -X POST localhost:8080/register -d '{"team": [], "icoAmount": -1}'
{
  "error":"ICO amount cannot be negative"
}
$ curl -X POST localhost:8080/register -d '{"team": [], "blockReward": -1}'
{
  "error":"block reward cannot be negative"
}
$ curl -X POST localhost:8080/register -d '{"team": []}'
{
  "error":"name cannot be null"
}
$ curl -X POST localhost:8080/register -d '{"name": "bitcoin", "symbol": "btc", "description": "The original cryptocurrency", "team": [], "icoAmount": 0, "blockReward": 12.5, "fundingStatus": "no-ico", "foundedDate": "2009-01-03", "coinType": "currency", "website": "https://bitcoin.org/en/"}'
{
  "id":"1"
}
$ curl -X POST localhost:8080/register -d '{"name": "bitcoin", "symbol": "btc", "description": "The original cryptocurrency", "team": [], "icoAmounockReward": 12.5, "fundingStatus": "no-ico", "foundedDate": "2009-01-03", "coinType": "currency", "website": "https://bitcoin.org/en/"}'
{
  "error":"symbol btc already exists"
}
$ curl -X POST localhost:8080/register -d '{"name": "ethereum", "symbol": "eth", "description": "The world computer", "team": ["Vitalik Buterin"], "icoAmount": 0, "blockReward": 3, "fundingStatus": "no-ico", "foundedDate": "2015-07-30", "coinType": "platform", "website": "https://www.ethereum.org/"}'
{
  "id":"2"
}
```
# Search examples
```
$ curl -X GET localhost:8080/search
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":12.5,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  },
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Vitalik Buterin"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X GET localhost:8080/search?name=bitcoin
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":12.5,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  }
]
$ curl -X GET localhost:8080/search?symbol=btc,eth
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":12.5,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  },
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Vitalik Buterin"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X GET localhost:8080/search?fundingStatus=pre-ico
[]
$ curl -X GET "localhost:8080/search?coinType=currency&coinType=platform"
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":12.5,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  },
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Vitalik Buterin"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X GET localhost:8080/search?startDate=2009-01-04
[
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Vitalik Buterin"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X GET localhost:8080/search?endDate=2009-01-04
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":12.5,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  },
]
$ curl -X GET localhost:8080/search?startDate=a
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":12.5,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  },
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Vitalik Buterin"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X GET localhost:8080/search?endDate=a
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":12.5,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  },
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Vitalik Buterin"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X GET "localhost:8080/search?symbol=neo&coinType=platform"
[]
$ curl -X GET "localhost:8080/search?fundingStatus=no-ico&coinType=currency,platform"
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":12.5,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  },
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Vitalik Buterin"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
```
# Update examples
```
$ curl -X POST localhost:8080/update -d '{'
false
$ curl -X POST localhost:8080/update -d '{}'
false
$ curl -X POST localhost:8080/update -d '{"foundedDate": "12/25/2017}'
false
$ curl -X POST localhost:8080/update -d '{"foundedDate": "2140-12-25}'
false
$ curl -X POST localhost:8080/update -d '{"team": [], "icoAmount": -1}'
false
$ curl -X POST localhost:8080/update -d '{"team": [], "blockReward": -1}'
false
$ curl -X POST localhost:8080/update -d '{"id": "2"}'
false
$ curl -X POST localhost:8080/update -d '{"id": "3", "description": "id 3 does not exist"}'
false
$ curl -X POST localhost:8080/update -d '{"id": "2", "team": ["Troll User"]}'
true
$ curl -X GET localhost:8080/search?symbol=eth
[
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Troll User"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X POST localhost:8080/update -d '{"id": "2", "team": []}'
true
$ curl -X GET localhost:8080/search?symbol=eth
[
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":[],
    "icoAmount":0,
    "blockReward":3
    ,"fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X POST localhost:8080/update -d '{"id": "2", "team": ["Vitalik Buterin"]}'
true
$ curl -X GET localhost:8080/search?symbol=eth
[
  {
    "id":"2",
    "name":"Ethereum",
    "symbol":"ETH",
    "description":"The world computer",
    "team":
      [
        "Vitalik Buterin"
      ],
    "icoAmount":0,
    "blockReward":3,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2015-07-30",
    "coinType":"Platform",
    "website":"https://www.ethereum.org/"
  }
]
$ curl -X POST localhost:8080/update -d '{"id": "1", "blockReward": 6.25}'
true
$ curl -X GET localhost:8080/search?symbol=btc
[
  {
    "id":"1",
    "name":"Bitcoin",
    "symbol":"BTC",
    "description":"The original cryptocurrency",
    "team":[],
    "icoAmount":0,
    "blockReward":6.25,
    "fundingStatus":"NO-ICO",
    "foundedDate":"2009-01-03",
    "coinType":"Currency",
    "website":"https://bitcoin.org/en/"
  },
]
```
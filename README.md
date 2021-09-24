# Key value storage

## How to run
go run main.go

## How to test
go test

## Simple API for key value storage with expiration timestamp.

### End point for saving data:
POST localhost:8080
 ` {
    key: string,
    data: []byte, // base64
    expiration_date: string
  }`


Example POST request body:
`{
    "Key": "key2",
    "Data": "bGtua1pubQ==",
    "Timestamp": "2022-11-01T22:08:41Z"
}`\
Example POST response body:
`{}`

###  End point for getting data
GET localhost:8080
  `{
    key: string
  }`

Example GET request body:
`{
    "key": "key2"
}`

Example GET resposne body:
`{
    "Key": "key2",
    "Data": "bGtua1pubQ==",
    "Timestamp": "2022-11-01T22:08:41Z",
    "CreateTime": "2021-09-22T22:14:49.6000244+02:00"
}`

Extra features:
1. JSON API

7. Data Expiration: automatically removes expired data

8. Data Deduplication: if we have 2 exactly same data with different key, then we save them only once, but we refer to them with 2 keys

9. Priority data: when multiple values under the same key, we return the last saved data

13. CI: Circle CI

12. Docker:
Requirements:
 - docker daemon running for `docker compose`

  - 1. a) Build Image for Docker:
  `docker build --tag docker-key-value-storage .`
  - 1. b) Run Image in docker: 
  `docker run --publish 8080:8080 docker-key-value-storage`
  - 2.  Build and run application in docker:
  `docker-compose up`


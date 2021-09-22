Key value storage

Simple API for key value storage with expiration timestamp.

End point for saving data:
POST localhost:8080
  {
    key: string,
    data: base64, //binary data
    expiration_date: timestamp
  }

End point for getting data
GET localhost:8080
  {
    key: string
  }

Extra features:
1. JSON API
2. ?WebSocket
3. ?Websocket Subscribe: subscribe to selected `key` and get information about changes for specific `key`
5. gRPC standart
7. Data Expiration: automatically removes expired data
8. Data Deduplication: if we have 2 exactly same data with different key, then we save them only once, but we refer to them with 2 `key`s
9. Priority data: when multiple values under the same key, we return the last saved data

13. ? CI: Circle CI or GitHub Actions


12. Docker:
Requirements:
 - docker daemon running for `docker compose`

Build Image for Docker:
  `docker build --tag docker-key-value-storage .`
Run Image in docker: 
  `docker run --publish 8080:8080 docker-key-value-storage`
Build and run application in docker:
  `docker-compose up`
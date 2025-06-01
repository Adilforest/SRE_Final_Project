module BikeStoreGolang/services/order-service

replace BikeStoreGolang/services/product-service => ../product-service

go 1.24.3

require (
	github.com/joho/godotenv v1.5.1
	github.com/nats-io/nats.go v1.42.0
	github.com/sirupsen/logrus v1.9.3
	go.mongodb.org/mongo-driver v1.17.3
	google.golang.org/grpc v1.72.1
	google.golang.org/protobuf v1.36.6
)

require (
	BikeStoreGolang/services/product-service v0.0.0-00010101000000-000000000000
	github.com/golang/snappy v0.0.4 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
)

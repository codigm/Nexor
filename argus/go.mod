module github.com/olympus/argus

go 1.21

require (
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.0.5
	github.com/segmentio/kafka-go v0.4.42
	github.com/olympus/titan v0.0.0
)

replace github.com/olympus/titan => ../titan

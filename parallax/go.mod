module github.com/nexor/parallax

go 1.21

require (
	github.com/lib/pq v1.10.9
	github.com/neo4j/neo4j-go-driver/v5 v5.24.0
	github.com/redis/go-redis/v9 v9.0.5
	github.com/segmentio/kafka-go v0.4.42
	github.com/nexor/flux v0.0.0
)

replace github.com/nexor/flux => ../flux

const { Kafka } = require('kafkajs');
const redis = require('redis');

async function main() {
  console.log("Initializing TITAN Analytics Engine...");

  const redisClient = redis.createClient({
    url: `redis://${process.env.REDIS_ADDR || 'localhost:6379'}`
  });
  await redisClient.connect();

  const brokers = (process.env.KAFKA_BROKERS || 'localhost:29092').split(',');
  const kafka = new Kafka({
    clientId: 'titan-core',
    brokers: brokers
  });

  const consumer = kafka.consumer({ groupId: 'titan-analytics-group' });
  await consumer.connect();
  await consumer.subscribe({ topic: 'transactions', fromBeginning: false });

  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      const payload = message.value.toString();
      console.log(`[TITAN] Aggregating analytics for: ${payload}`);
      // Add analytics logic here
    },
  });
}

main().catch(console.error);

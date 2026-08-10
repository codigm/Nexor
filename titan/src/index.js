const { Kafka } = require('kafkajs');
const redis = require('redis');
const express = require('express');
const http = require('http');
const { Server } = require('socket.io');
const cors = require('cors');

async function main() {
  console.log("Initializing TITAN Analytics Engine with WebSockets...");

  // Setup Express & Socket.IO
  const app = express();
  app.use(cors());
  const server = http.createServer(app);
  const io = new Server(server, {
    cors: {
      origin: "*", // For development
      methods: ["GET", "POST"]
    }
  });

  io.on('connection', (socket) => {
    console.log(`[TITAN] Client connected: ${socket.id}`);
    socket.on('disconnect', () => {
      console.log(`[TITAN] Client disconnected: ${socket.id}`);
    });
  });

  // Setup Redis
  const redisClient = redis.createClient({
    url: `redis://${process.env.REDIS_ADDR || 'localhost:6379'}`
  });
  await redisClient.connect();

  // Setup Kafka Consumer
  const brokers = (process.env.KAFKA_BROKERS || 'localhost:29092').split(',');
  const kafka = new Kafka({
    clientId: 'titan-core',
    brokers: brokers
  });

  const consumer = kafka.consumer({ groupId: 'titan-analytics-group' });
  await consumer.connect();
  await consumer.subscribe({ topic: 'payment-initiated', fromBeginning: false });

  // Expose an HTTP endpoint just for health checks
  app.get('/health', (req, res) => res.send({ status: 'OK' }));

  // Start HTTP + WebSocket Server
  const PORT = process.env.PORT || 4000;
  server.listen(PORT, () => {
    console.log(`[TITAN] Server listening on port ${PORT}`);
  });

  // Consume Kafka and Broadcast
  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      const payload = message.value.toString();
      console.log(`[TITAN] Received transaction: ${payload}`);
      
      try {
        const tx = JSON.parse(payload);
        // Broadcast to all connected websocket clients
        io.emit('transaction', tx);
      } catch (err) {
        console.error("[TITAN] Failed to parse Kafka message", err);
      }
    },
  });
}

main().catch(console.error);

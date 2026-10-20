const { Kafka } = require('kafkajs');
const redis = require('redis');
const express = require('express');
const http = require('http');
const { Server } = require('socket.io');
const cors = require('cors');

async function main() {
  console.log("Initializing FLUX Analytics Engine with WebSockets...");

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
    console.log(`[FLUX] Client connected: ${socket.id}`);
    socket.on('disconnect', () => {
      console.log(`[FLUX] Client disconnected: ${socket.id}`);
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
    clientId: 'flux-core',
    brokers: brokers
  });

  const consumer = kafka.consumer({ groupId: 'flux-analytics-group' });
  await consumer.connect();
  await consumer.subscribe({ topic: 'payment-initiated', fromBeginning: false });

  // Expose an HTTP endpoint just for health checks
  app.get('/health', (req, res) => res.send({ status: 'OK' }));

  // Start HTTP + WebSocket Server
  const PORT = process.env.PORT || 4000;
  server.listen(PORT, () => {
    console.log(`[FLUX] Server listening on port ${PORT}`);
  });

  // Consume Kafka and Broadcast
  await consumer.run({
    eachMessage: async ({ topic, partition, message }) => {
      const payload = message.value.toString();
      console.log(`[FLUX] Received transaction: ${payload}`);
      
      try {
        const tx = JSON.parse(payload);
        // Broadcast to all connected websocket clients
        io.emit('transaction', tx);
      } catch (err) {
        console.error("[FLUX] Failed to parse Kafka message", err);
      }
    },
  });
}

main().catch(console.error);

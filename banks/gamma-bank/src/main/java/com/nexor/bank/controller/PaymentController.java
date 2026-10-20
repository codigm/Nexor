package com.nexor.bank.controller;

import com.nexor.common.dto.PaymentRequest;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.time.Instant;
import java.util.UUID;

@Slf4j
@CrossOrigin(origins = "*")
@RestController
@RequestMapping("/api/transfer")
@RequiredArgsConstructor
public class PaymentController {

    private final KafkaTemplate<String, PaymentRequest> kafkaTemplate;
    private static final String TOPIC = "payment-initiated";

    @PostMapping
    public ResponseEntity<String> initiateTransfer(@RequestBody PaymentRequest request) {
        if (request.getTransactionId() == null) {
            request.setTransactionId(UUID.randomUUID());
        }
        request.setTimestamp(Instant.now());
        
        log.info("VesperBank initiating transfer: {} from {} to {} for {} {}", 
                request.getTransactionId(), request.getSenderAccount(), 
                request.getReceiverAccount(), request.getAmount(), request.getCurrency());
                
        kafkaTemplate.send(TOPIC, request.getTransactionId().toString(), request);
        
        return ResponseEntity.accepted().body("Transfer initiated with ID: " + request.getTransactionId());
    }
}

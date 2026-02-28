package com.kubesec.transaction.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kubesec.transaction.model.dto.TransactionEvent;
import io.nats.client.Connection;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

@Service
public class NatsPublisher {

    private static final Logger log = LoggerFactory.getLogger(NatsPublisher.class);

    private final Connection natsConnection;
    private final ObjectMapper objectMapper;

    public NatsPublisher(Connection natsConnection, ObjectMapper objectMapper) {
        this.natsConnection = natsConnection;
        this.objectMapper = objectMapper;
    }

    public void publishTransactionCompleted(TransactionEvent event) {
        try {
            byte[] data = objectMapper.writeValueAsBytes(event);
            natsConnection.publish("transactions.completed", data);
        } catch (Exception e) {
            log.warn("Failed to publish event: {}", e.getMessage());
        }
    }
}

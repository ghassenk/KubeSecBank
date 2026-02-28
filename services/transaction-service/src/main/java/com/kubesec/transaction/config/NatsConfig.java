package com.kubesec.transaction.config;

import io.nats.client.Connection;
import io.nats.client.Nats;
import io.nats.client.Options;
import jakarta.annotation.PreDestroy;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Profile;

import java.io.IOException;

@Configuration
@Profile("!test")
public class NatsConfig {

    private static final Logger log = LoggerFactory.getLogger(NatsConfig.class);
    private Connection connection;

    @Bean
    public Connection natsConnection(AppConfig appConfig) throws IOException, InterruptedException {
        Options options = new Options.Builder()
                .server(appConfig.getNatsUrl())
                .build();
        connection = Nats.connect(options);
        log.info("Connected to NATS at {}", appConfig.getNatsUrl());
        return connection;
    }

    @PreDestroy
    public void destroy() {
        if (connection != null) {
            try {
                connection.drain(java.time.Duration.ofSeconds(5));
                log.info("NATS connection drained");
            } catch (Exception e) {
                log.warn("Error draining NATS connection: {}", e.getMessage());
                try {
                    connection.close();
                } catch (InterruptedException ex) {
                    Thread.currentThread().interrupt();
                }
            }
        }
    }
}

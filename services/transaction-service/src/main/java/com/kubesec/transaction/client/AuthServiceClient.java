package com.kubesec.transaction.client;

import com.kubesec.transaction.config.AppConfig;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import java.util.Map;

@Component
public class AuthServiceClient {

    private final RestClient restClient;

    public AuthServiceClient(AppConfig config) {
        this.restClient = RestClient.builder()
                .baseUrl(config.getAuthServiceUrl())
                .build();
    }

    public ValidateResponse validateToken(String token) {
        return restClient.post()
                .uri("/api/v1/auth/validate")
                .body(Map.of("token", token))
                .retrieve()
                .body(ValidateResponse.class);
    }

    public record ValidateResponse(boolean valid, String user_id) {}
}

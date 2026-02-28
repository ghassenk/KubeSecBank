package com.kubesec.transaction.client;

import com.kubesec.transaction.config.AppConfig;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import java.math.BigDecimal;
import java.util.UUID;

@Component
public class AccountServiceClient {

    private final RestClient restClient;

    public AccountServiceClient(AppConfig config) {
        this.restClient = RestClient.builder()
                .baseUrl(config.getAccountServiceUrl())
                .build();
    }

    public BalanceResponse getBalance(UUID accountId, String authHeader) {
        return restClient.get()
                .uri("/accounts/{id}/balance", accountId)
                .header("Authorization", authHeader)
                .retrieve()
                .body(BalanceResponse.class);
    }

    public record BalanceResponse(UUID account_id, BigDecimal balance, String currency) {}
}

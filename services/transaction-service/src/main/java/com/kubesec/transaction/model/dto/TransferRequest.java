package com.kubesec.transaction.model.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.math.BigDecimal;
import java.util.UUID;

public record TransferRequest(
        @JsonProperty("from_account_id") UUID fromAccountId,
        @JsonProperty("to_account_id") UUID toAccountId,
        BigDecimal amount,
        String currency,
        String description
) {}

package com.kubesec.transaction.model.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.util.UUID;

public record TransactionEvent(
        @JsonProperty("transaction_id") UUID transactionId,
        @JsonProperty("from_account_id") UUID fromAccountId,
        @JsonProperty("to_account_id") UUID toAccountId,
        BigDecimal amount,
        String currency,
        String type,
        String status,
        OffsetDateTime timestamp
) {}

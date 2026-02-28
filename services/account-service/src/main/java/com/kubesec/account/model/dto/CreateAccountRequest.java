package com.kubesec.account.model.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

public record CreateAccountRequest(
        @JsonProperty("user_id") String userId,
        @JsonProperty("account_type") String accountType,
        String currency
) {}

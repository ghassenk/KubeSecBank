package com.kubesec.account.model.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

public record CreateUserRequest(
        String email,
        @JsonProperty("full_name") String fullName
) {}

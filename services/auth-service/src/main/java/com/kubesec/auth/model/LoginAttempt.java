package com.kubesec.auth.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.time.OffsetDateTime;

public record LoginAttempt(
        String id,
        String email,
        boolean success,
        @JsonProperty("ip_address") String ipAddress,
        @JsonProperty("created_at") OffsetDateTime createdAt
) {}

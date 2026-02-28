package com.kubesec.auth.model.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonInclude(JsonInclude.Include.NON_NULL)
public record TokenValidationResponse(
        boolean valid,
        @JsonProperty("user_id") String userId,
        String email
) {
    public static TokenValidationResponse invalid() {
        return new TokenValidationResponse(false, null, null);
    }
}

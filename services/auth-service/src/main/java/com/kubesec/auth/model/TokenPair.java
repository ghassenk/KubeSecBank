package com.kubesec.auth.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public record TokenPair(
        @JsonProperty("access_token") String accessToken,
        @JsonProperty("refresh_token") String refreshToken
) {}

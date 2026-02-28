package com.kubesec.auth.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.time.OffsetDateTime;

public class Session {

    private String id;

    @JsonProperty("user_id")
    private String userId;

    private String token;

    @JsonProperty("expires_at")
    private OffsetDateTime expiresAt;

    @JsonProperty("created_at")
    private OffsetDateTime createdAt;

    public Session() {}

    public Session(String id, String userId, String token, OffsetDateTime expiresAt, OffsetDateTime createdAt) {
        this.id = id;
        this.userId = userId;
        this.token = token;
        this.expiresAt = expiresAt;
        this.createdAt = createdAt;
    }

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getUserId() { return userId; }
    public void setUserId(String userId) { this.userId = userId; }

    public String getToken() { return token; }
    public void setToken(String token) { this.token = token; }

    public OffsetDateTime getExpiresAt() { return expiresAt; }
    public void setExpiresAt(OffsetDateTime expiresAt) { this.expiresAt = expiresAt; }

    public OffsetDateTime getCreatedAt() { return createdAt; }
    public void setCreatedAt(OffsetDateTime createdAt) { this.createdAt = createdAt; }
}

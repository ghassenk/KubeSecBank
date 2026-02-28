package com.kubesec.auth.repository;

import com.kubesec.auth.model.LoginAttempt;
import com.kubesec.auth.model.Session;

import java.time.Duration;
import java.time.OffsetDateTime;

public interface AuthRepository {

    // Session operations (PostgreSQL)
    void createSession(Session session);
    void deleteSession(String token);
    void deleteSessionsByUserId(String userId);

    // Login attempt operations (PostgreSQL)
    void recordLoginAttempt(LoginAttempt attempt);
    int getRecentFailedAttempts(String email, OffsetDateTime since);

    // Token blacklist (Redis)
    void blacklistToken(String token, Duration expiry);
    boolean isTokenBlacklisted(String token);

    // Session cache (Redis)
    void cacheSession(String token, String userId, Duration expiry);
    void invalidateCachedSession(String token);
}

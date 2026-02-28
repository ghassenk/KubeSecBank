package com.kubesec.auth.repository;

import com.kubesec.auth.model.LoginAttempt;
import com.kubesec.auth.model.Session;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

import java.time.Duration;
import java.time.OffsetDateTime;

@Repository
public class AuthRepositoryImpl implements AuthRepository {

    private static final String BLACKLIST_PREFIX = "blacklist:";
    private static final String SESSION_CACHE_PREFIX = "session:";

    private final JdbcTemplate jdbc;
    private final StringRedisTemplate redis;

    public AuthRepositoryImpl(JdbcTemplate jdbc, StringRedisTemplate redis) {
        this.jdbc = jdbc;
        this.redis = redis;
    }

    // --- Session operations (PostgreSQL) ---

    @Override
    public void createSession(Session session) {
        jdbc.update(
                "INSERT INTO sessions (id, user_id, token, expires_at, created_at) VALUES (?, ?, ?, ?, ?)",
                session.getId(), session.getUserId(), session.getToken(),
                session.getExpiresAt(), session.getCreatedAt()
        );
    }

    @Override
    public void deleteSession(String token) {
        jdbc.update("DELETE FROM sessions WHERE token = ?", token);
    }

    @Override
    public void deleteSessionsByUserId(String userId) {
        jdbc.update("DELETE FROM sessions WHERE user_id = ?", userId);
    }

    // --- Login attempt operations (PostgreSQL) ---

    @Override
    public void recordLoginAttempt(LoginAttempt attempt) {
        jdbc.update(
                "INSERT INTO login_attempts (id, email, success, ip_address, created_at) VALUES (?, ?, ?, ?, ?)",
                attempt.id(), attempt.email(), attempt.success(),
                attempt.ipAddress(), attempt.createdAt()
        );
    }

    @Override
    public int getRecentFailedAttempts(String email, OffsetDateTime since) {
        Integer count = jdbc.queryForObject(
                "SELECT COUNT(*) FROM login_attempts WHERE email = ? AND success = false AND created_at > ?",
                Integer.class, email, since
        );
        return count != null ? count : 0;
    }

    // --- Token blacklist (Redis) ---

    @Override
    public void blacklistToken(String token, Duration expiry) {
        redis.opsForValue().set(BLACKLIST_PREFIX + token, "1", expiry);
    }

    @Override
    public boolean isTokenBlacklisted(String token) {
        return Boolean.TRUE.equals(redis.hasKey(BLACKLIST_PREFIX + token));
    }

    // --- Session cache (Redis) ---

    @Override
    public void cacheSession(String token, String userId, Duration expiry) {
        redis.opsForValue().set(SESSION_CACHE_PREFIX + token, userId, expiry);
    }

    @Override
    public void invalidateCachedSession(String token) {
        redis.delete(SESSION_CACHE_PREFIX + token);
    }
}

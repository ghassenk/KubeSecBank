package com.kubesec.auth.service;

import com.kubesec.auth.model.LoginAttempt;
import com.kubesec.auth.model.Session;
import com.kubesec.auth.model.TokenPair;
import com.kubesec.auth.model.dto.TokenValidationResponse;
import com.kubesec.auth.repository.AuthRepository;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.JwtException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.UUID;

@Service
public class AuthService {

    private static final Logger log = LoggerFactory.getLogger(AuthService.class);

    private final AuthRepository repository;
    private final JwtService jwtService;

    public AuthService(AuthRepository repository, JwtService jwtService) {
        this.repository = repository;
        this.jwtService = jwtService;
    }

    public TokenPair login(String email, String password, String ipAddress) {
        OffsetDateTime now = OffsetDateTime.now(ZoneOffset.UTC);

        // Check for brute-force attempts
        int failedCount = repository.getRecentFailedAttempts(email, now.minusMinutes(15));
        if (failedCount >= 5) {
            throw new RateLimitedException("too many failed login attempts, try again later");
        }

        // TODO: Replace with real user lookup via account-service
        String userId = "user-" + email;
        boolean authenticated = true;

        // Record login attempt
        LoginAttempt attempt = new LoginAttempt(
                UUID.randomUUID().toString(),
                email,
                authenticated,
                ipAddress,
                now
        );
        try {
            repository.recordLoginAttempt(attempt);
        } catch (Exception e) {
            log.error("error recording login attempt: {}", e.getMessage());
        }

        if (!authenticated) {
            throw new AuthenticationException("invalid credentials");
        }

        // Issue tokens
        TokenPair tokenPair = jwtService.issueTokens(userId, email);

        // Persist session
        Session session = new Session(
                UUID.randomUUID().toString(),
                userId,
                tokenPair.accessToken(),
                now.plus(jwtService.getAccessTokenExpiry()),
                now
        );
        try {
            repository.createSession(session);
            repository.cacheSession(session.getToken(), session.getUserId(), jwtService.getAccessTokenExpiry());
        } catch (Exception e) {
            log.error("error creating session: {}", e.getMessage());
            throw new RuntimeException("failed to create session");
        }

        return tokenPair;
    }

    public void logout(String token, String userId) {
        try {
            repository.blacklistToken(token, jwtService.getAccessTokenExpiry());
        } catch (Exception e) {
            log.error("error blacklisting token: {}", e.getMessage());
        }
        try {
            repository.deleteSession(token);
        } catch (Exception e) {
            log.error("error deleting session: {}", e.getMessage());
        }
        repository.invalidateCachedSession(token);
        log.info("user {} logged out", userId);
    }

    public TokenPair refresh(String refreshToken) {
        // Check if blacklisted
        if (repository.isTokenBlacklisted(refreshToken)) {
            throw new AuthenticationException("token has been revoked");
        }

        // Validate refresh token
        Claims claims;
        try {
            claims = jwtService.parseToken(refreshToken);
        } catch (JwtException e) {
            throw new AuthenticationException("invalid or expired refresh token");
        }

        String tokenType = claims.get("type", String.class);
        if (!"refresh".equals(tokenType)) {
            throw new AuthenticationException("not a refresh token");
        }

        String userId = claims.get("user_id", String.class);
        String email = claims.get("email", String.class);

        // Blacklist old refresh token
        repository.blacklistToken(refreshToken, jwtService.getRefreshTokenExpiry());

        // Issue new pair
        return jwtService.issueTokens(userId, email);
    }

    public TokenValidationResponse validate(String token) {
        // Check blacklist
        if (repository.isTokenBlacklisted(token)) {
            return TokenValidationResponse.invalid();
        }

        try {
            Claims claims = jwtService.parseToken(token);
            String userId = claims.get("user_id", String.class);
            String email = claims.get("email", String.class);
            return new TokenValidationResponse(true, userId, email);
        } catch (JwtException e) {
            return TokenValidationResponse.invalid();
        }
    }

    // Custom exceptions
    public static class AuthenticationException extends RuntimeException {
        public AuthenticationException(String message) { super(message); }
    }

    public static class RateLimitedException extends RuntimeException {
        public RateLimitedException(String message) { super(message); }
    }
}

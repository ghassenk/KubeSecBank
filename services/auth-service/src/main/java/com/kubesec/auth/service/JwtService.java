package com.kubesec.auth.service;

import com.kubesec.auth.config.AppConfig;
import com.kubesec.auth.model.TokenPair;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.JwtException;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import org.springframework.stereotype.Service;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.time.Instant;
import java.util.Date;
import java.util.Map;

@Service
public class JwtService {

    private final SecretKey key;
    private final Duration accessTokenExpiry;
    private static final Duration REFRESH_TOKEN_EXPIRY = Duration.ofDays(7);

    public JwtService(AppConfig config) {
        this.key = Keys.hmacShaKeyFor(config.getJwtSecret().getBytes(StandardCharsets.UTF_8));
        this.accessTokenExpiry = config.getJwtExpiryDuration();
    }

    public TokenPair issueTokens(String userId, String email) {
        Instant now = Instant.now();

        String accessToken = Jwts.builder()
                .claims(Map.of("user_id", userId, "email", email, "type", "access"))
                .issuedAt(Date.from(now))
                .expiration(Date.from(now.plus(accessTokenExpiry)))
                .signWith(key)
                .compact();

        String refreshToken = Jwts.builder()
                .claims(Map.of("user_id", userId, "email", email, "type", "refresh"))
                .issuedAt(Date.from(now))
                .expiration(Date.from(now.plus(REFRESH_TOKEN_EXPIRY)))
                .signWith(key)
                .compact();

        return new TokenPair(accessToken, refreshToken);
    }

    public Claims parseToken(String token) throws JwtException {
        return Jwts.parser()
                .verifyWith(key)
                .build()
                .parseSignedClaims(token)
                .getPayload();
    }

    public Duration getAccessTokenExpiry() {
        return accessTokenExpiry;
    }

    public Duration getRefreshTokenExpiry() {
        return REFRESH_TOKEN_EXPIRY;
    }
}

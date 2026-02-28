package com.kubesec.auth.controller;

import com.kubesec.auth.model.Credentials;
import com.kubesec.auth.model.TokenPair;
import com.kubesec.auth.model.dto.RefreshRequest;
import com.kubesec.auth.model.dto.TokenValidationResponse;
import com.kubesec.auth.model.dto.ValidateRequest;
import com.kubesec.auth.service.AuthService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
public class AuthController {

    private final AuthService authService;

    public AuthController(AuthService authService) {
        this.authService = authService;
    }

    @GetMapping("/healthz")
    public Map<String, String> health() {
        return Map.of("status", "ok");
    }

    @PostMapping("/api/v1/auth/login")
    public TokenPair login(@RequestBody Credentials credentials, HttpServletRequest request) {
        if (credentials.email() == null || credentials.email().isEmpty()
                || credentials.password() == null || credentials.password().isEmpty()) {
            throw new IllegalArgumentException("email and password are required");
        }
        return authService.login(credentials.email(), credentials.password(), request.getRemoteAddr());
    }

    @PostMapping("/api/v1/auth/logout")
    public Map<String, String> logout(HttpServletRequest request) {
        String authHeader = request.getHeader("Authorization");
        if (authHeader == null || !authHeader.startsWith("Bearer ")) {
            throw new AuthService.AuthenticationException("unauthorized");
        }
        String token = authHeader.substring(7);

        // Get userId from request attribute (set by JwtAuthFilter)
        String userId = (String) request.getAttribute("userId");

        authService.logout(token, userId);
        return Map.of("message", "logged out successfully");
    }

    @PostMapping("/api/v1/auth/refresh")
    public TokenPair refresh(@RequestBody RefreshRequest request) {
        if (request.refreshToken() == null || request.refreshToken().isEmpty()) {
            throw new IllegalArgumentException("refresh_token is required");
        }
        return authService.refresh(request.refreshToken());
    }

    @PostMapping("/api/v1/auth/validate")
    public TokenValidationResponse validate(@RequestBody ValidateRequest request) {
        if (request.token() == null || request.token().isEmpty()) {
            throw new IllegalArgumentException("token is required");
        }
        return authService.validate(request.token());
    }
}

package com.kubesec.auth.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

import java.time.Duration;

@Configuration
@ConfigurationProperties(prefix = "app")
public class AppConfig {

    private String jwtSecret = "change-me-in-production";
    private int jwtExpiry = 15; // minutes

    public String getJwtSecret() { return jwtSecret; }
    public void setJwtSecret(String jwtSecret) { this.jwtSecret = jwtSecret; }

    public int getJwtExpiry() { return jwtExpiry; }
    public void setJwtExpiry(int jwtExpiry) { this.jwtExpiry = jwtExpiry; }

    public Duration getJwtExpiryDuration() {
        return Duration.ofMinutes(jwtExpiry);
    }
}

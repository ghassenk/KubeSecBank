package com.kubesec.transaction.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Configuration
@ConfigurationProperties(prefix = "app")
public class AppConfig {

    private String natsUrl = "nats://localhost:4222";
    private String authServiceUrl = "http://localhost:8082";
    private String accountServiceUrl = "http://localhost:8081";

    public String getNatsUrl() { return natsUrl; }
    public void setNatsUrl(String natsUrl) { this.natsUrl = natsUrl; }

    public String getAuthServiceUrl() { return authServiceUrl; }
    public void setAuthServiceUrl(String authServiceUrl) { this.authServiceUrl = authServiceUrl; }

    public String getAccountServiceUrl() { return accountServiceUrl; }
    public void setAccountServiceUrl(String accountServiceUrl) { this.accountServiceUrl = accountServiceUrl; }
}

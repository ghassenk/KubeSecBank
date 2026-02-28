package com.kubesec.auth;

import org.junit.jupiter.api.Test;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.context.TestConfiguration;
import org.springframework.context.annotation.Bean;
import org.springframework.data.redis.connection.RedisConnectionFactory;
import org.springframework.data.redis.connection.lettuce.LettuceConnectionFactory;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.test.context.TestPropertySource;

@SpringBootTest
@TestPropertySource(properties = {
        "spring.datasource.url=jdbc:h2:mem:testdb",
        "spring.datasource.driver-class-name=org.h2.Driver",
        "spring.flyway.enabled=false",
        "spring.autoconfigure.exclude=org.springframework.boot.autoconfigure.data.redis.RedisAutoConfiguration,org.springframework.boot.autoconfigure.data.redis.RedisRepositoriesAutoConfiguration",
        "app.jwt-secret=test-secret-key-that-is-at-least-256-bits-long-for-hs256"
})
class ApplicationTest {

    @TestConfiguration
    static class TestRedisConfig {
        @Bean
        RedisConnectionFactory redisConnectionFactory() {
            LettuceConnectionFactory factory = new LettuceConnectionFactory("localhost", 6379);
            factory.setValidateConnection(false);
            return factory;
        }

        @Bean
        StringRedisTemplate stringRedisTemplate(RedisConnectionFactory factory) {
            return new StringRedisTemplate(factory);
        }
    }

    @Test
    void contextLoads() {
    }
}

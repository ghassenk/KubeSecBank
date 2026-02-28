package com.kubesec.auth.filter;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;

@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
public class RateLimitFilter extends OncePerRequestFilter {

    private static final int LIMIT = 60;
    private static final long WINDOW_SECONDS = 60;

    private final ConcurrentHashMap<String, List<Instant>> requests = new ConcurrentHashMap<>();

    @Override
    protected boolean shouldNotFilter(HttpServletRequest request) {
        String path = request.getRequestURI();
        // Only rate-limit auth endpoints
        return !path.startsWith("/api/v1/auth/");
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response,
                                    FilterChain chain) throws ServletException, IOException {
        String ip = request.getRemoteAddr();
        Instant now = Instant.now();
        Instant windowStart = now.minusSeconds(WINDOW_SECONDS);

        List<Instant> timestamps = requests.compute(ip, (key, existing) -> {
            List<Instant> valid = new ArrayList<>();
            if (existing != null) {
                for (Instant t : existing) {
                    if (t.isAfter(windowStart)) {
                        valid.add(t);
                    }
                }
            }
            return valid;
        });

        synchronized (timestamps) {
            if (timestamps.size() >= LIMIT) {
                response.setContentType("application/json");
                response.setStatus(429);
                response.getWriter().write("{\"error\":\"rate limit exceeded\"}");
                return;
            }
            timestamps.add(now);
        }

        chain.doFilter(request, response);
    }
}

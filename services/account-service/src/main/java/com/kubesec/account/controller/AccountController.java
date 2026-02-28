package com.kubesec.account.controller;

import com.kubesec.account.model.Account;
import com.kubesec.account.model.User;
import com.kubesec.account.model.dto.CreateAccountRequest;
import com.kubesec.account.model.dto.CreateUserRequest;
import com.kubesec.account.service.AccountService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
public class AccountController {

    private final AccountService accountService;

    public AccountController(AccountService accountService) {
        this.accountService = accountService;
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "ok");
    }

    @PostMapping("/api/v1/users")
    public ResponseEntity<User> createUser(@RequestBody CreateUserRequest request) {
        if (request.email() == null || request.email().isEmpty()
                || request.fullName() == null || request.fullName().isEmpty()) {
            throw new IllegalArgumentException("email and full_name are required");
        }
        User user = accountService.createUser(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(user);
    }

    @GetMapping("/api/v1/users/{id}")
    public User getUser(@PathVariable UUID id) {
        return accountService.getUser(id);
    }

    @GetMapping("/api/v1/users/{id}/accounts")
    public List<Account> listAccountsByUser(@PathVariable UUID id) {
        return accountService.listAccountsByUser(id);
    }

    @PostMapping("/api/v1/accounts")
    public ResponseEntity<Account> createAccount(@RequestBody CreateAccountRequest request) {
        Account account = accountService.createAccount(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(account);
    }

    @GetMapping("/api/v1/accounts/{id}")
    public Account getAccount(@PathVariable UUID id) {
        return accountService.getAccount(id);
    }
}

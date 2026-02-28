package com.kubesec.account.service;

import com.kubesec.account.exception.ResourceNotFoundException;
import com.kubesec.account.model.Account;
import com.kubesec.account.model.User;
import com.kubesec.account.model.dto.CreateAccountRequest;
import com.kubesec.account.model.dto.CreateUserRequest;
import com.kubesec.account.repository.AccountRepository;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.UUID;

@Service
public class AccountService {

    private final AccountRepository repository;

    public AccountService(AccountRepository repository) {
        this.repository = repository;
    }

    public User createUser(CreateUserRequest request) {
        OffsetDateTime now = OffsetDateTime.now(ZoneOffset.UTC);
        User user = new User(
                UUID.randomUUID(),
                request.email(),
                request.fullName(),
                "pending",
                now,
                now
        );
        repository.createUser(user);
        return user;
    }

    public User getUser(UUID id) {
        return repository.getUser(id)
                .orElseThrow(() -> new ResourceNotFoundException("user not found"));
    }

    public Account createAccount(CreateAccountRequest request) {
        UUID userId = UUID.fromString(request.userId());

        String accountType = request.accountType();
        if (!"checking".equals(accountType) && !"savings".equals(accountType)) {
            throw new IllegalArgumentException("account_type must be checking or savings");
        }

        String currency = request.currency();
        if (currency == null || currency.isEmpty()) {
            currency = "USD";
        }

        OffsetDateTime now = OffsetDateTime.now(ZoneOffset.UTC);
        Account account = new Account(
                UUID.randomUUID(),
                userId,
                accountType,
                BigDecimal.ZERO,
                currency,
                "active",
                now,
                now
        );
        repository.createAccount(account);
        return account;
    }

    public Account getAccount(UUID id) {
        return repository.getAccount(id)
                .orElseThrow(() -> new ResourceNotFoundException("account not found"));
    }

    public List<Account> listAccountsByUser(UUID userId) {
        return repository.listAccountsByUser(userId);
    }
}

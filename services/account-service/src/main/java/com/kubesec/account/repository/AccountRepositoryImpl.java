package com.kubesec.account.repository;

import com.kubesec.account.model.Account;
import com.kubesec.account.model.User;
import org.springframework.dao.EmptyResultDataAccessException;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.stereotype.Repository;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public class AccountRepositoryImpl implements AccountRepository {

    private final JdbcTemplate jdbc;

    public AccountRepositoryImpl(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    @Override
    public void createUser(User user) {
        jdbc.update(
                "INSERT INTO users (id, email, full_name, kyc_status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
                user.getId(), user.getEmail(), user.getFullName(),
                user.getKycStatus(), user.getCreatedAt(), user.getUpdatedAt()
        );
    }

    @Override
    public Optional<User> getUser(UUID id) {
        try {
            return Optional.ofNullable(jdbc.queryForObject(
                    "SELECT id, email, full_name, kyc_status, created_at, updated_at FROM users WHERE id = ?",
                    this::mapUser, id
            ));
        } catch (EmptyResultDataAccessException e) {
            return Optional.empty();
        }
    }

    @Override
    public void createAccount(Account account) {
        jdbc.update(
                "INSERT INTO accounts (id, user_id, account_type, balance, currency, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
                account.getId(), account.getUserId(), account.getAccountType(),
                account.getBalance(), account.getCurrency(), account.getStatus(),
                account.getCreatedAt(), account.getUpdatedAt()
        );
    }

    @Override
    public Optional<Account> getAccount(UUID id) {
        try {
            return Optional.ofNullable(jdbc.queryForObject(
                    "SELECT id, user_id, account_type, balance, currency, status, created_at, updated_at FROM accounts WHERE id = ?",
                    this::mapAccount, id
            ));
        } catch (EmptyResultDataAccessException e) {
            return Optional.empty();
        }
    }

    @Override
    public List<Account> listAccountsByUser(UUID userId) {
        return jdbc.query(
                "SELECT id, user_id, account_type, balance, currency, status, created_at, updated_at FROM accounts WHERE user_id = ? ORDER BY created_at",
                this::mapAccount, userId
        );
    }

    private User mapUser(ResultSet rs, int rowNum) throws SQLException {
        return new User(
                rs.getObject("id", UUID.class),
                rs.getString("email"),
                rs.getString("full_name"),
                rs.getString("kyc_status"),
                rs.getObject("created_at", java.time.OffsetDateTime.class),
                rs.getObject("updated_at", java.time.OffsetDateTime.class)
        );
    }

    private Account mapAccount(ResultSet rs, int rowNum) throws SQLException {
        return new Account(
                rs.getObject("id", UUID.class),
                rs.getObject("user_id", UUID.class),
                rs.getString("account_type"),
                rs.getBigDecimal("balance"),
                rs.getString("currency"),
                rs.getString("status"),
                rs.getObject("created_at", java.time.OffsetDateTime.class),
                rs.getObject("updated_at", java.time.OffsetDateTime.class)
        );
    }
}

package com.kubesec.transaction.repository;

import com.kubesec.transaction.model.Transaction;
import com.kubesec.transaction.model.TransactionFilter;
import org.springframework.dao.EmptyResultDataAccessException;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;
import org.springframework.transaction.annotation.Transactional;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public class TransactionRepositoryImpl implements TransactionRepository {

    private final JdbcTemplate jdbc;

    public TransactionRepositoryImpl(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    @Override
    @Transactional
    public void create(Transaction txn) {
        jdbc.update(
                "INSERT INTO transactions (id, from_account_id, to_account_id, amount, currency, type, status, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
                txn.getId(), txn.getFromAccountId(), txn.getToAccountId(),
                txn.getAmount(), txn.getCurrency(), txn.getType(), txn.getStatus(),
                txn.getDescription(), txn.getCreatedAt(), txn.getUpdatedAt()
        );
    }

    @Override
    public Optional<Transaction> getById(UUID id) {
        try {
            return Optional.ofNullable(jdbc.queryForObject(
                    "SELECT id, from_account_id, to_account_id, amount, currency, type, status, description, created_at, updated_at FROM transactions WHERE id = ?",
                    this::mapTransaction, id
            ));
        } catch (EmptyResultDataAccessException e) {
            return Optional.empty();
        }
    }

    @Override
    public List<Transaction> list(TransactionFilter filter) {
        StringBuilder query = new StringBuilder(
                "SELECT id, from_account_id, to_account_id, amount, currency, type, status, description, created_at, updated_at FROM transactions WHERE 1=1"
        );
        List<Object> args = new ArrayList<>();

        if (filter.getAccountId() != null) {
            query.append(" AND (from_account_id = ? OR to_account_id = ?)");
            args.add(filter.getAccountId());
            args.add(filter.getAccountId());
        }

        if (filter.getStatus() != null && !filter.getStatus().isEmpty()) {
            query.append(" AND status = ?");
            args.add(filter.getStatus());
        }

        query.append(" ORDER BY created_at DESC");

        if (filter.getLimit() > 0) {
            query.append(" LIMIT ?");
            args.add(filter.getLimit());
        }

        if (filter.getOffset() > 0) {
            query.append(" OFFSET ?");
            args.add(filter.getOffset());
        }

        return jdbc.query(query.toString(), this::mapTransaction, args.toArray());
    }

    @Override
    public void updateStatus(UUID id, String status) {
        int rows = jdbc.update(
                "UPDATE transactions SET status = ?, updated_at = NOW() WHERE id = ?",
                status, id
        );
        if (rows == 0) {
            throw new IllegalStateException("transaction " + id + " not found");
        }
    }

    private Transaction mapTransaction(ResultSet rs, int rowNum) throws SQLException {
        return new Transaction(
                rs.getObject("id", UUID.class),
                rs.getObject("from_account_id", UUID.class),
                rs.getObject("to_account_id", UUID.class),
                rs.getBigDecimal("amount"),
                rs.getString("currency"),
                rs.getString("type"),
                rs.getString("status"),
                rs.getString("description"),
                rs.getObject("created_at", java.time.OffsetDateTime.class),
                rs.getObject("updated_at", java.time.OffsetDateTime.class)
        );
    }
}

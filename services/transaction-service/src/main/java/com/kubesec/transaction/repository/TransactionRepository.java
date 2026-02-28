package com.kubesec.transaction.repository;

import com.kubesec.transaction.model.Transaction;
import com.kubesec.transaction.model.TransactionFilter;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

public interface TransactionRepository {

    void create(Transaction transaction);

    Optional<Transaction> getById(UUID id);

    List<Transaction> list(TransactionFilter filter);

    void updateStatus(UUID id, String status);
}

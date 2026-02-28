package com.kubesec.transaction.service;

import com.kubesec.transaction.client.AccountServiceClient;
import com.kubesec.transaction.exception.InsufficientBalanceException;
import com.kubesec.transaction.exception.ResourceNotFoundException;
import com.kubesec.transaction.model.Transaction;
import com.kubesec.transaction.model.TransactionFilter;
import com.kubesec.transaction.model.dto.TransactionEvent;
import com.kubesec.transaction.model.dto.TransferRequest;
import com.kubesec.transaction.repository.TransactionRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.lang.Nullable;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.time.OffsetDateTime;
import java.time.ZoneOffset;
import java.util.List;
import java.util.UUID;

@Service
public class TransactionService {

    private static final Logger log = LoggerFactory.getLogger(TransactionService.class);

    private final TransactionRepository repository;
    private final AccountServiceClient accountClient;
    private final NatsPublisher natsPublisher;

    public TransactionService(TransactionRepository repository,
                              AccountServiceClient accountClient,
                              @Nullable NatsPublisher natsPublisher) {
        this.repository = repository;
        this.accountClient = accountClient;
        this.natsPublisher = natsPublisher;
    }

    public Transaction createTransfer(TransferRequest request, String authHeader) {
        // Validate
        if (request.fromAccountId() == null || request.toAccountId() == null) {
            throw new IllegalArgumentException("from_account_id and to_account_id are required");
        }
        if (request.amount() == null || request.amount().compareTo(BigDecimal.ZERO) <= 0) {
            throw new IllegalArgumentException("amount must be positive");
        }
        if (request.currency() == null || request.currency().isEmpty()) {
            throw new IllegalArgumentException("currency is required");
        }
        if (request.fromAccountId().equals(request.toAccountId())) {
            throw new IllegalArgumentException("cannot transfer to the same account");
        }

        // Check balance via account-service
        AccountServiceClient.BalanceResponse balance;
        try {
            balance = accountClient.getBalance(request.fromAccountId(), authHeader);
        } catch (Exception e) {
            log.error("ERROR: check balance: {}", e.getMessage());
            throw new RuntimeException("could not verify account balance");
        }

        if (balance.balance().compareTo(request.amount()) < 0) {
            throw new InsufficientBalanceException("insufficient balance");
        }

        // Create transaction
        OffsetDateTime now = OffsetDateTime.now(ZoneOffset.UTC);
        Transaction txn = new Transaction(
                UUID.randomUUID(),
                request.fromAccountId(),
                request.toAccountId(),
                request.amount(),
                request.currency(),
                "transfer",
                "pending",
                request.description() != null ? request.description() : "",
                now,
                now
        );

        repository.create(txn);

        // Mark completed
        txn.setStatus("completed");
        txn.setUpdatedAt(OffsetDateTime.now(ZoneOffset.UTC));
        try {
            repository.updateStatus(txn.getId(), "completed");
        } catch (Exception e) {
            log.error("ERROR: update status: {}", e.getMessage());
        }

        // Publish event
        TransactionEvent event = new TransactionEvent(
                txn.getId(), txn.getFromAccountId(), txn.getToAccountId(),
                txn.getAmount(), txn.getCurrency(), txn.getType(),
                txn.getStatus(), txn.getUpdatedAt()
        );
        if (natsPublisher != null) {
            natsPublisher.publishTransactionCompleted(event);
        }

        return txn;
    }

    public Transaction getTransaction(UUID id) {
        return repository.getById(id)
                .orElseThrow(() -> new ResourceNotFoundException("transaction not found"));
    }

    public List<Transaction> listTransactions(TransactionFilter filter) {
        return repository.list(filter);
    }
}

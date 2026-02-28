package com.kubesec.transaction.controller;

import com.kubesec.transaction.model.Transaction;
import com.kubesec.transaction.model.TransactionFilter;
import com.kubesec.transaction.model.dto.TransferRequest;
import com.kubesec.transaction.service.TransactionService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

@RestController
public class TransactionController {

    private final TransactionService transactionService;

    public TransactionController(TransactionService transactionService) {
        this.transactionService = transactionService;
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "healthy");
    }

    @PostMapping("/transactions/transfer")
    public ResponseEntity<Transaction> createTransfer(@RequestBody TransferRequest request,
                                                       HttpServletRequest httpRequest) {
        String authHeader = httpRequest.getHeader("Authorization");
        Transaction txn = transactionService.createTransfer(request, authHeader);
        return ResponseEntity.status(HttpStatus.CREATED).body(txn);
    }

    @GetMapping("/transactions/{id}")
    public Transaction getTransaction(@PathVariable UUID id) {
        return transactionService.getTransaction(id);
    }

    @GetMapping("/transactions")
    public Map<String, Object> listTransactions(
            @RequestParam(name = "account_id", required = false) UUID accountId,
            @RequestParam(required = false) String status,
            @RequestParam(required = false, defaultValue = "20") int limit,
            @RequestParam(required = false, defaultValue = "0") int offset) {

        if (limit < 1 || limit > 100) limit = 20;
        if (offset < 0) offset = 0;

        TransactionFilter filter = new TransactionFilter();
        filter.setAccountId(accountId);
        filter.setStatus(status);
        filter.setLimit(limit);
        filter.setOffset(offset);

        List<Transaction> transactions = transactionService.listTransactions(filter);

        Map<String, Object> response = new LinkedHashMap<>();
        response.put("transactions", transactions);
        response.put("limit", filter.getLimit());
        response.put("offset", filter.getOffset());
        return response;
    }
}

package com.kubesec.transaction.model;

import java.util.UUID;

public class TransactionFilter {

    private UUID accountId;
    private String status;
    private int limit = 20;
    private int offset = 0;

    public UUID getAccountId() { return accountId; }
    public void setAccountId(UUID accountId) { this.accountId = accountId; }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }

    public int getLimit() { return limit; }
    public void setLimit(int limit) { this.limit = limit; }

    public int getOffset() { return offset; }
    public void setOffset(int offset) { this.offset = offset; }
}

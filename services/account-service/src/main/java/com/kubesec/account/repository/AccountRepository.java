package com.kubesec.account.repository;

import com.kubesec.account.model.Account;
import com.kubesec.account.model.User;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

public interface AccountRepository {

    void createUser(User user);

    Optional<User> getUser(UUID id);

    void createAccount(Account account);

    Optional<Account> getAccount(UUID id);

    List<Account> listAccountsByUser(UUID userId);
}

# Scenario

You're designing a large banking system that enables users to transfer money
from their accounts to other accounts using the following method. Each account
has an entry in a table in a database that looks something like:

id | first_name | last_name | balance
12 | lucas     | bn        | 150

Since there are so many different accounts, you have split your database into
multiple shards. To decide which shard to place data on, we hash the account ID
and apply a modulo 2 operation. If the result is zero we place the data on the
first database, and otherwise we place the data on the second database.




# Setting up the databases

### Setting up postgres
```bash
brew install postgresql
```

```bash
brew services start postgresql
```

```bash
psql -d postgres
```

```
CREATE TABLE account (
    id TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    balance INTEGER NOT NULL
);
```
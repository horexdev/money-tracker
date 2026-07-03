# Source Postgres Exporter

`cmd/source-exporter` writes a PostgreSQL snapshot into `MoneyTrackerBackup` JSON v1 for the Android restore pipeline.

Run:

```bash
go run ./cmd/source-exporter \
  --database-url "$DATABASE_URL" \
  --output moneytracker.mtbackup.json
```

To export one source user only, use `--user-id` as an operator filter:

```bash
go run ./cmd/source-exporter \
  --database-url "$DATABASE_URL" \
  --user-id 123456789 \
  --output moneytracker-user.mtbackup.json
```

The generated file contains neutral local profiles, accounts, categories, transactions, transfers, budgets, recurring transactions, savings goals, goal transactions, exchange rate snapshots, and transaction templates.

Identity boundary:

- `users.id`, usernames, first names, last names, Telegram init data, bot/chat metadata, admin data, source database IDs, and `legacy_*` fields are not serialized.
- `--user-id` only limits which source rows are read.
- Relationships are rebuilt with export-local refs such as `account:a000001` and `transaction:t000001`.
- Refs are generated per export and are not copied from PostgreSQL primary keys.

The exporter writes plain JSON for migration tooling. Store the output as sensitive financial data and remove temporary files after import.

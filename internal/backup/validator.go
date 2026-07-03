package backup

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	refPattern        = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:[-:.][a-z0-9_]+)*$`)
	forbiddenRefParts = []string{"telegram", "source", "legacy", "init_data", "bot", "chat"}
)

func ValidateBackup(backup MoneyTrackerBackup) error {
	if backup.Format != MoneyTrackerBackupFormat {
		return fmt.Errorf("unsupported backup format: %s", backup.Format)
	}
	if backup.Version != MoneyTrackerBackupVersion {
		return fmt.Errorf("unsupported backup version: %d", backup.Version)
	}

	if _, err := uniqueRefs("profiles", profileRefs(backup.Profiles)); err != nil {
		return err
	}
	for _, profile := range backup.Profiles {
		if err := validateProfile(profile); err != nil {
			return err
		}
	}
	return nil
}

func validateProfile(profile BackupProfile) error {
	accountRefs, err := uniqueRefs(profile.Ref+".accounts", accountRefs(profile.Accounts))
	if err != nil {
		return err
	}
	categoryRefs, err := uniqueRefs(profile.Ref+".categories", categoryRefs(profile.Categories))
	if err != nil {
		return err
	}
	transactionRefs, err := uniqueRefs(profile.Ref+".transactions", transactionRefs(profile.Transactions))
	if err != nil {
		return err
	}
	goalRefs, err := uniqueRefs(profile.Ref+".savings_goals", goalRefs(profile.SavingsGoals))
	if err != nil {
		return err
	}

	if _, err := uniqueRefs(profile.Ref+".transfers", transferRefs(profile.Transfers)); err != nil {
		return err
	}
	if _, err := uniqueRefs(profile.Ref+".budgets", budgetRefs(profile.Budgets)); err != nil {
		return err
	}
	if _, err := uniqueRefs(profile.Ref+".recurring_transactions", recurringRefs(profile.RecurringTransactions)); err != nil {
		return err
	}
	if _, err := uniqueRefs(profile.Ref+".goal_transactions", goalTransactionRefs(profile.GoalTransactions)); err != nil {
		return err
	}
	if _, err := uniqueRefs(profile.Ref+".exchange_rate_overrides", exchangeRateOverrideRefs(profile.ExchangeRateOverrides)); err != nil {
		return err
	}
	if _, err := uniqueRefs(profile.Ref+".transaction_templates", transactionTemplateRefs(profile.TransactionTemplates)); err != nil {
		return err
	}

	for _, transaction := range profile.Transactions {
		if err := requireResolvedRef(transaction.AccountRef, accountRefs, profile.Ref+".transactions.account_ref"); err != nil {
			return err
		}
		if err := requireResolvedRef(transaction.CategoryRef, categoryRefs, profile.Ref+".transactions.category_ref"); err != nil {
			return err
		}
	}
	for _, transfer := range profile.Transfers {
		if err := requireResolvedRef(transfer.FromAccountRef, accountRefs, profile.Ref+".transfers.from_account_ref"); err != nil {
			return err
		}
		if err := requireResolvedRef(transfer.ToAccountRef, accountRefs, profile.Ref+".transfers.to_account_ref"); err != nil {
			return err
		}
		if transfer.FromTransactionRef != nil {
			if err := requireResolvedRef(*transfer.FromTransactionRef, transactionRefs, profile.Ref+".transfers.from_transaction_ref"); err != nil {
				return err
			}
		}
		if transfer.ToTransactionRef != nil {
			if err := requireResolvedRef(*transfer.ToTransactionRef, transactionRefs, profile.Ref+".transfers.to_transaction_ref"); err != nil {
				return err
			}
		}
	}
	for _, budget := range profile.Budgets {
		if err := requireResolvedRef(budget.CategoryRef, categoryRefs, profile.Ref+".budgets.category_ref"); err != nil {
			return err
		}
	}
	for _, recurring := range profile.RecurringTransactions {
		if err := requireResolvedRef(recurring.AccountRef, accountRefs, profile.Ref+".recurring_transactions.account_ref"); err != nil {
			return err
		}
		if err := requireResolvedRef(recurring.CategoryRef, categoryRefs, profile.Ref+".recurring_transactions.category_ref"); err != nil {
			return err
		}
	}
	for _, goal := range profile.SavingsGoals {
		if goal.AccountRef != nil {
			if err := requireResolvedRef(*goal.AccountRef, accountRefs, profile.Ref+".savings_goals.account_ref"); err != nil {
				return err
			}
		}
	}
	for _, transaction := range profile.GoalTransactions {
		if err := requireResolvedRef(transaction.GoalRef, goalRefs, profile.Ref+".goal_transactions.goal_ref"); err != nil {
			return err
		}
	}
	for _, template := range profile.TransactionTemplates {
		if err := requireResolvedRef(template.AccountRef, accountRefs, profile.Ref+".transaction_templates.account_ref"); err != nil {
			return err
		}
		if err := requireResolvedRef(template.CategoryRef, categoryRefs, profile.Ref+".transaction_templates.category_ref"); err != nil {
			return err
		}
	}
	return nil
}

func uniqueRefs(scope string, refs []string) (map[string]struct{}, error) {
	seen := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		if err := requireValidRef(scope, ref); err != nil {
			return nil, err
		}
		if _, ok := seen[ref]; ok {
			return nil, fmt.Errorf("duplicate export ref %q in %s", ref, scope)
		}
		seen[ref] = struct{}{}
	}
	return seen, nil
}

func requireValidRef(scope, ref string) error {
	if !refPattern.MatchString(ref) {
		return fmt.Errorf("invalid export ref %q in %s", ref, scope)
	}
	normalized := strings.ToLower(ref)
	for _, forbidden := range forbiddenRefParts {
		if strings.Contains(normalized, forbidden) {
			return fmt.Errorf("forbidden export ref part %q in %s", forbidden, scope)
		}
	}
	return nil
}

func requireResolvedRef(ref string, knownRefs map[string]struct{}, scope string) error {
	if _, ok := knownRefs[ref]; !ok {
		return fmt.Errorf("unresolved export ref %q in %s", ref, scope)
	}
	return nil
}

func profileRefs(values []BackupProfile) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func accountRefs(values []BackupAccount) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func categoryRefs(values []BackupCategory) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func transactionRefs(values []BackupTransaction) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func transferRefs(values []BackupTransfer) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func budgetRefs(values []BackupBudget) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func recurringRefs(values []BackupRecurringTransaction) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func goalRefs(values []BackupSavingsGoal) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func goalTransactionRefs(values []BackupGoalTransaction) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func exchangeRateOverrideRefs(values []BackupExchangeRateOverride) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

func transactionTemplateRefs(values []BackupTransactionTemplate) []string {
	refs := make([]string, 0, len(values))
	for _, value := range values {
		refs = append(refs, value.Ref)
	}
	return refs
}

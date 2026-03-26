// Package fsm implements a Redis-backed finite state machine for
// managing multi-step Telegram conversation flows.
package fsm

import "fmt"

// State represents the current step in a conversation flow.
type State string

const (
	// StateNone means the user has no active conversation flow.
	StateNone State = ""

	// Expense flow
	StateExpenseWaitAmount   State = "expense:wait_amount"
	StateExpenseWaitCategory State = "expense:wait_category"
	StateExpenseWaitNote     State = "expense:wait_note"

	// Income flow
	StateIncomeWaitAmount   State = "income:wait_amount"
	StateIncomeWaitCategory State = "income:wait_category"
	StateIncomeWaitNote     State = "income:wait_note"

	// Stats flow
	StateStatsWaitPeriod State = "stats:wait_period"
)

// stateKey returns the Redis key for storing a user's current FSM state.
func stateKey(userID int64) string {
	return fmt.Sprintf("fsm:state:%d", userID)
}

// dataKey returns the Redis key for storing intermediate FSM data.
func dataKey(userID int64, field string) string {
	return fmt.Sprintf("fsm:data:%d:%s", userID, field)
}

// dataPattern returns the Redis key pattern for all data keys of a user.
func dataPattern(userID int64) string {
	return fmt.Sprintf("fsm:data:%d:*", userID)
}

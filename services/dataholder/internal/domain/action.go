package domain

import "time"

// Action is an action that the rules permit now, with each arrangement that it ends.
type Action[A ~string] struct {
	Action A
	Ends   []ArrangementState
}

type AccountAction string

const (
	AccountActionClose             AccountAction = "close"
	AccountActionCreateTransaction AccountAction = "create-transaction"
)

type ArrangementAction string

const ArrangementActionWithdraw ArrangementAction = "withdraw"

// IneligibilityEnds gives each arrangement that a loss of eligibility at the time ends.
func IneligibilityEnds(arrangements []ArrangementState, at time.Time) []ArrangementState {
	ends := []ArrangementState{}

	for _, state := range arrangements {
		if !state.Ended(at) {
			ends = append(ends, state)
		}
	}

	return ends
}

// PermittedAccountActions gives the actions of each account by its AccountID, with the rules of Account.Close and
// Account.AcceptTransaction. A close of the last open account ends the arrangements that IneligibilityEnds gives.
func PermittedAccountActions(
	accounts []Account,
	arrangements []ArrangementState,
	now time.Time,
) map[string][]Action[AccountAction] {
	last := len(OpenAccounts(accounts)) == 1
	actions := make(map[string][]Action[AccountAction], len(accounts))

	for _, account := range accounts {
		permitted := []Action[AccountAction]{}

		if _, err := account.Close(); err == nil {
			ends := []ArrangementState{}
			if last {
				ends = IneligibilityEnds(arrangements, now)
			}

			permitted = append(permitted, Action[AccountAction]{Action: AccountActionClose, Ends: ends})
		}

		if account.AcceptTransaction() == nil {
			permitted = append(
				permitted, Action[AccountAction]{Action: AccountActionCreateTransaction, Ends: []ArrangementState{}},
			)
		}

		actions[account.AccountID] = permitted
	}

	return actions
}

// PermittedArrangementActions gives the actions of each arrangement by its ArrangementID, with the rules of
// WithdrawArrangement.
func PermittedArrangementActions(
	arrangements []ArrangementState,
	now time.Time,
) map[string][]Action[ArrangementAction] {
	actions := make(map[string][]Action[ArrangementAction], len(arrangements))

	for _, state := range arrangements {
		actions[state.Arrangement.ArrangementID] = []Action[ArrangementAction]{}
		if state.Ended(now) {
			continue
		}

		actions[state.Arrangement.ArrangementID] = []Action[ArrangementAction]{
			{Action: ArrangementActionWithdraw, Ends: []ArrangementState{state}},
		}
	}

	return actions
}

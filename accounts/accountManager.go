package accounts

import (
	"Flygon/db"
	"Flygon/pogo"
	"errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/guregu/null.v4"
	"sync"
	"time"
)

type AccountDetails struct {
	Username string
	Password string
}

type AccountProvider interface {
	GetNextAccount(testAccount func(a db.Account) bool) *AccountDetails
	ReleaseAccount(account AccountDetails)
	MarkSuspended(username string)
	MarkBanned(username string)
	MarkDisabled(username string)
	UpdateDetailsFromGame(username string, fromGame pogo.GetPlayerOutProto)
}

type AccountManager struct {
	//	accountNo   int
	accounts    []db.Account
	inUse       []bool
	db          db.DbDetails
	accountLock sync.Mutex
}

type AccountStatus struct {
	DbRow db.Account
	InUse bool
}

func (a *AccountManager) GetAccountDetails() []AccountStatus {
	accountStatusList := make([]AccountStatus, 0)
	for n, account := range a.accounts {
		accountStatusList = append(accountStatusList, AccountStatus{
			DbRow: account,
			InUse: a.inUse[n],
		})
	}

	return accountStatusList
}

func (a *AccountManager) LoadAccounts(dbDetails db.DbDetails) {
	var err error
	a.accounts, err = db.GetAccountRecords(dbDetails)
	if err != nil {
		panic(err)
	}
	a.inUse = make([]bool, len(a.accounts))
	a.db = dbDetails
}

func (a *AccountManager) ReloadAccounts() {
	a.accountLock.Lock()
	defer a.accountLock.Unlock()

	accounts, err := db.GetAccountRecords(a.db)
	if err != nil {
		log.Errorf("Error reloading accounts: %s", err)
		return
	}

	foundRecords := make([]bool, len(a.accounts))
	for _, account := range accounts {
		found := false
		for x := 0; x < len(a.accounts); x++ {
			if a.accounts[x].Username == account.Username {
				found = true
				a.accounts[x] = account
				foundRecords[x] = true
				break
			}
		}

		if !found {
			log.Infof("Found new account %s", account.Username)
			a.accounts = append(a.accounts, account)
			a.inUse = append(a.inUse, false)
			foundRecords = append(foundRecords, true)
		}
	}

	for x := 0; x < len(a.accounts); x++ {
		if !foundRecords[x] {
			log.Infof("Account %s no longer exists", a.accounts[x].Username)
			a.accounts = append(a.accounts[:x], a.accounts[x+1:]...)
			a.inUse = append(a.inUse[:x], a.inUse[x+1:]...)
			foundRecords = append(foundRecords[:x], foundRecords[x+1:]...)
		}
	}
}

func (a *AccountManager) GetNextAccount(testAccount func(a db.Account) bool) *AccountDetails {
	// Find the least recently used account

	a.accountLock.Lock()
	defer a.accountLock.Unlock()

	time24hrAgo := time.Now().Add(-24 * time.Hour).Unix()
	timeNow := time.Now().Unix()

	minReleased := int64(0)
	bestAccount := -1
	for x := 0; x < len(a.accounts); x++ {
		account := a.accounts[x]

		if a.inUse[x] ||
			account.Suspended ||
			account.Banned ||
			int64(account.WarnExpiration) > timeNow ||
			(account.LastDisabled.Valid && account.LastDisabled.Int64 > time24hrAgo) {
			continue
		}

		if !testAccount(account) {
			continue
		}

		lastReleased := account.LastReleased.ValueOrZero()
		if lastReleased == 0 {
			continue // this shouldn't happen because of the inuse check above
		}

		if minReleased == 0 || lastReleased < minReleased {
			minReleased = lastReleased
			bestAccount = x
		}
	}

	if bestAccount == -1 {
		return nil // We could not find an account :-(
	}

	account := &a.accounts[bestAccount]
	if minReleased > time24hrAgo {
		log.Warnf("Selected account %s was last released %d minutes ago, which is less than 24 hours ago. This is probably not what you want.", account.Username, (time.Now().Unix()-minReleased)/60)
	}

	a.inUse[bestAccount] = true
	account.LastReleased = null.NewInt(0, false)
	account.LastSelected = null.IntFrom(time.Now().Unix())
	if err := db.MarkSelected(a.db, account.Username); err != nil {
		log.Errorf("Error marking account %s as selected: %s", account.Username, err)
	}

	return &AccountDetails{
		Username: account.Username,
		Password: account.Password,
	}
}

func (a *AccountManager) ReleaseAccount(account AccountDetails) {
	for x := range a.accounts {
		if a.accounts[x].Username == account.Username {
			a.inUse[x] = false
			a.accounts[x].LastReleased = null.IntFrom(time.Now().Unix())
		}
	}

	if err := db.MarkReleased(a.db, account.Username); err != nil {
		log.Errorf("Error marking account %s as released: %s", account.Username, err)
	}
}

// IsValidAccount This should be called on every job request - expensive?
func (a *AccountManager) IsValidAccount(username string) (bool, error) {
	for x := range a.accounts {
		if a.accounts[x].Username == username {
			time24hrAgo := time.Now().Add(-24 * time.Hour).Unix()
			timeNow := time.Now().Unix()
			return a.accounts[x].Suspended ||
				a.accounts[x].Banned ||
				int64(a.accounts[x].WarnExpiration) > timeNow ||
				(a.accounts[x].LastDisabled.Valid && a.accounts[x].LastDisabled.Int64 > time24hrAgo), nil
		}
	}
	log.Errorf("Account with username '%s' not found in accounts", username)
	return false, errors.New("account " + username + " not found")
}

func (a *AccountManager) MarkWarned(username string) {
	for x := range a.accounts {
		if a.accounts[x].Username == username {
			a.accounts[x].Warn = true
			a.accounts[x].WarnExpiration = int(time.Now().Unix() + 7*24*60*60)
		}
	}

	if err := db.MarkWarned(a.db, username); err != nil {
		log.Errorf("Error marking account %s as warned: %s", username, err)
	}
}

func (a *AccountManager) MarkSuspended(username string) {
	for x := range a.accounts {
		if a.accounts[x].Username == username {
			a.accounts[x].Suspended = true
		}
	}

	if err := db.MarkSuspended(a.db, username); err != nil {
		log.Errorf("Error marking account %s as suspended: %s", username, err)
	}
}

func (a *AccountManager) MarkBanned(username string) {
	for x := range a.accounts {
		if a.accounts[x].Username == username {
			a.accounts[x].Banned = true
		}
	}
	if err := db.MarkBanned(a.db, username); err != nil {
		log.Errorf("Error marking account %s as banned: %s", username, err)
	}
}

func (a *AccountManager) MarkDisabled(username string) {
	for x := range a.accounts {
		if a.accounts[x].Username == username {
			a.accounts[x].Disabled = true
			a.accounts[x].LastDisabled = null.IntFrom(time.Now().Unix())
		}
	}
	if err := db.MarkDisabled(a.db, username); err != nil {
		log.Errorf("Error marking account %s as disabled: %s", username, err)
	}
}

func (a *AccountManager) MarkTutorialDone(username string) {
	for x := range a.accounts {
		if a.accounts[x].Username == username {
			level := a.accounts[x].Level
			if level == 0 {
				a.accounts[x].Level = 1
			}
		}
	}
	if err := db.MarkTutorialDone(a.db, username); err != nil {
		log.Errorf("Error marking account %s as tutorial done: %s", username, err)
	}
}

func (a *AccountManager) AccountExists(username string) bool {
	for x := range a.accounts {
		if a.accounts[x].Username == username {
			return true
		}
	}
	return false
}

func (a *AccountManager) UpdateDetailsFromGame(username string, fromGame pogo.GetPlayerOutProto, trainerlevel int) {
	for x := range a.accounts {
		if a.accounts[x].Username == username {
			dbRecord := a.accounts[x]
			if dbRecord.Warn != fromGame.Warn ||
				dbRecord.WarnExpiration != int(fromGame.WarnExpireMs) ||
				dbRecord.Suspended != (fromGame.WasSuspended && !fromGame.SuspendedMessageAcknowledged) ||
				(trainerlevel > 0 && dbRecord.Level != trainerlevel) {
				a.accounts[x].Suspended = fromGame.WasSuspended && !fromGame.SuspendedMessageAcknowledged
				a.accounts[x].Warn = fromGame.Warn
				a.accounts[x].WarnExpiration = int(fromGame.WarnExpireMs) / 1000
				if trainerlevel > 0 {
					a.accounts[x].Level = trainerlevel
				}
				db.UpdateDetailsFromGame(a.db, username, fromGame, trainerlevel)
			}
			break
		}
	}
}

func SelectLevel30(account db.Account) bool {
	return account.Level >= 30
}

func SelectUnderLevel30(account db.Account) bool {
	return account.Level < 30
}

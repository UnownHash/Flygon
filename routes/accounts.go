package routes

import (
	"flygon/accounts"
	"flygon/db"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"net/http"
	"time"
)

type ApiAccountRecord struct {
	Id                           string `json:"id"`
	Username                     string `json:"username"`
	Password                     string `json:"password"`
	InUse                        bool   `json:"in_use"`
	Level                        int    `json:"level"`
	Suspended                    bool   `json:"suspended"`
	Banned                       bool   `json:"banned"`
	Invalid                      bool   `json:"invalid"`
	Warn                         bool   `json:"warn"`
	Disabled                     bool   `json:"disabled"`
	LastDisabled                 *int64 `json:"last_disabled"`
	LastBanned                   *int64 `json:"last_banned"`
	LastSuspended                *int64 `json:"last_suspended"`
	WarnMessageAcknowledged      bool   `json:"warn_message_acknowledged"`
	SuspendedMessageAcknowledged bool   `json:"suspended_message_acknowledged"`
	WarnExpirationMs             int    `json:"warn_expiration_ms"`
}

func accountStatusToApiAccount(account accounts.AccountStatus) ApiAccountRecord {
	time24HoursAgo := time.Now().Add(-24 * time.Hour).Unix()
	return ApiAccountRecord{
		Username:                     account.DbRow.Username,
		Id:                           account.DbRow.Username,
		Password:                     account.DbRow.Password,
		InUse:                        account.InUse,
		Level:                        account.DbRow.Level,
		Suspended:                    account.DbRow.Suspended,
		Banned:                       account.DbRow.Banned,
		Invalid:                      account.DbRow.Invalid,
		Warn:                         account.DbRow.Warn,
		Disabled:                     account.DbRow.LastDisabled.ValueOrZero() > time24HoursAgo,
		LastDisabled:                 account.DbRow.LastDisabled.Ptr(),
		LastBanned:                   account.DbRow.LastBanned.Ptr(),
		LastSuspended:                account.DbRow.LastSuspended.Ptr(),
		WarnMessageAcknowledged:      false,
		SuspendedMessageAcknowledged: false,
		WarnExpirationMs:             account.DbRow.WarnExpiration,
	}
}

func buildAccountResponse() []ApiAccountRecord {
	accounts := accountManager.GetAccountDetails()

	accountRecords := []ApiAccountRecord{}

	for _, a := range accounts {
		accountRecords = append(accountRecords,
			accountStatusToApiAccount(a))
	}

	return accountRecords
}

func GetAccounts(c *gin.Context) {
	//authHeader := c.Request.Header.Get("X-Golbat-Secret")
	//if config.Config.ApiSecret != "" {
	//	if authHeader != config.Config.ApiSecret {
	//		log.Errorf("ClearQuests: Incorrect authorisation received (%s)", authHeader)
	//		c.String(http.StatusUnauthorized, "Unauthorised")
	//		return
	//	}
	//}

	accountRecords := buildAccountResponse()
	paginateAndSort(c, accountRecords)
}

func GetOneAccount(context *gin.Context) {
	accountName := context.Param("account_name")

	if accountName == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid account name"})
		return
	}

	accountsDetails := accountManager.GetAccountDetails()
	index := slices.IndexFunc(accountsDetails, func(account accounts.AccountStatus) bool {
		return account.DbRow.Username == accountName
	})

	if index == -1 {
		context.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	context.JSON(http.StatusOK, accountStatusToApiAccount(accountsDetails[index]))
}

func GetAccountsStats(context *gin.Context) {
	stats, statsErr := db.GetAccountsStats(*dbDetails)

	if statsErr != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": statsErr.Error()})
		return
	}

	context.JSON(http.StatusOK, stats)
}

func GetLevelStats(context *gin.Context) {
	stats, statsErr := db.GetLevelStats(*dbDetails)

	if statsErr != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": statsErr.Error()})
		return
	}

	context.JSON(http.StatusOK, stats)
}

type ApiSimpleAccountRecord struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ApiNewAccountBatch struct {
	Accounts     []ApiSimpleAccountRecord `json:"accounts"`
	DefaultLevel int                      `json:"default_level"`
}

func PostAccount(c *gin.Context) {
	var requestBody ApiNewAccountBatch

	if err := c.BindJSON(&requestBody); err != nil {
		log.Warnf("POST /accounts/ Error during post area %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	var insertedNo int64 = 0

	newAccounts := make([]db.NewAccountRow, len(requestBody.Accounts))

	for i, account := range requestBody.Accounts {
		newAccounts[i] = db.NewAccountRow{
			Username: account.Username,
			Password: account.Password,
			Level:    requestBody.DefaultLevel,
		}
	}
	res, err := db.InsertBulkAccounts(*dbDetails, newAccounts)
	if err == nil {
		insertedNo = res
	}

	accountManager.ReloadAccounts()

	c.JSON(http.StatusAccepted, gin.H{"updated": insertedNo})
}

type ApiDeleteAccountBatch struct {
	Usernames []string `json:"usernames"`
}

func DeleteAccount(c *gin.Context) {
	var requestBody ApiNewAccountBatch

	if err := c.BindJSON(&requestBody); err != nil {
		// DO SOMETHING WITH THE ERROR
	}

	c.JSON(http.StatusNotImplemented, gin.H{"message": "Accounts deletion not implemented yet"})
}

type ApiPatchAccountBatch struct {
	Accounts []ApiPatchAccountSingle `json:"accounts"`
}

type ApiPatchAccountSingle struct {
	Username    string `json:"username"`
	NewUsername string `json:"new_username"`
	NewPassword string `json:"new_password"`
}

func PatchAccount(c *gin.Context) {
	var requestBody ApiPatchAccountBatch

	if err := c.BindJSON(&requestBody); err != nil {
		// DO SOMETHING WITH THE ERROR
	}

	c.JSON(http.StatusNotImplemented, gin.H{"message": "Accounts updates not implemented yet"})
}

func GetReloadAccounts(c *gin.Context) {
	accountManager.ReloadAccounts()
	c.Status(http.StatusOK)
}

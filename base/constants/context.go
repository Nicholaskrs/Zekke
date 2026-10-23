package constants

import "time"

const (
	TransactionTimeOut         = 5 * time.Second // Used to set timeout of transaction in seconds (2 seconds)
	ContextAdminID             = "adminID"
	ContextUserID              = "ID"
	ContextAdminPrivilegeGroup = "privilegeGroup"
	AccessTTL                  = 15 * time.Minute
	RefreshTTL                 = 30 * 24 * time.Hour
	Timezone                   = "Asia/Jakarta"
)

package model

type Sync struct {
	Root
	IsFullSynced      bool  `json:"is_full_synced"`
	LastOnixSyncDate  int64 `json:"last_onix_sync_date"`
	LastAnnotSyncDate int64 `json:"last_annot_sync_date"`
}

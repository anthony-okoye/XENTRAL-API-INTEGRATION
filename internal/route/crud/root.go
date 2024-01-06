package crud

import (
	"bookbox-backend/internal/model"
	_ "embed"
	"encoding/json"
	"log"
	"sync"
)

var (
	//go:embed permissions.json
	PermissionsRaw []byte

	Permissions map[string]map[string]map[string]struct{}

	//used when editing and ordering products
	productMutex sync.Mutex
)

func IsAuthorized(issuer *model.User, operation string, entity string) (isAuth bool) {
	if issuer.Role == "admin" {
		return true
	}

	crudRoutes := Permissions[issuer.Role]
	allowedModels := crudRoutes[operation]
	_, isAuth = allowedModels[entity]
	return
}

func init() {
	err := json.Unmarshal(PermissionsRaw, &Permissions)
	if err != nil {
		log.Fatal("failed to parse permissions")
	}
}

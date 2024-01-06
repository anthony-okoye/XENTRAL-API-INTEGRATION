package postrun

import (
	"bookbox-backend/internal/cache"
	"encoding/base64"
	"fmt"
)

func Cache(raw []byte, entity string, data any) {

	input := base64.StdEncoding.EncodeToString(raw)
	cache.DataCache.Set(fmt.Sprintf(entity+":"+input), data)
}

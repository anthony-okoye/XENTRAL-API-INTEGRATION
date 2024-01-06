package prerun

import (
	"bookbox-backend/internal/cache"
	"bookbox-backend/internal/request"
	"encoding/base64"
	"fmt"
)

func ReadCached(raw []byte, inputRequest request.GetRequest) (data interface{}, found bool) {

	input := base64.StdEncoding.EncodeToString(raw)
	data, found = cache.DataCache.Get(fmt.Sprintf(inputRequest.Entity + ":" + input))

	return
}

package fbcodec

import (
	"fmt"

	"github.com/kigongo-vincent/my-broker-backend/fbs/gen/mybroker"
)

func OpenRequest(data []byte) (*mybroker.RequestEnv, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("empty body")
	}
	return mybroker.GetRootAsRequestEnv(data, 0), nil
}

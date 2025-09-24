package instance

import "multi-client-whatsapp/internal/types"

var Manager *types.InstanceManager

func InitializeManager() {
	Manager = &types.InstanceManager{
		Instances: make(map[string]*types.Instance),
	}
}

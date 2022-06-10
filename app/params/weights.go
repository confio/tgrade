package params

// Default simulation operation weights for messages and gov proposals
const (
	DefaultWeightMsgCreateValidator int = 100
	DefaultWeightMsgUpdateValidator int = 5
	DefaultWeightMsgDelegate        int = 100
	DefaultWeightMsgUndelegate      int = 90

	DefaultWeightMsgStoreCode           int = 50
	DefaultWeightMsgInstantiateContract int = 100
	DefaultWeightMsgExecuteContract     int = 100
)

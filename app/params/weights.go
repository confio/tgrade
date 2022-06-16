package params

// Default simulation operation weights for messages and gov proposals
const (
	DefaultWeightMsgCreateValidator int = 100
	DefaultWeightMsgUpdateValidator int = 10
	DefaultWeightMsgDelegate        int = 200
	DefaultWeightMsgUndelegate      int = 50

	DefaultWeightMsgStoreCode           int = 50
	DefaultWeightMsgInstantiateContract int = 100
	DefaultWeightMsgExecuteContract     int = 100
)

package enum

type ImageType string

const (
	// TODO NK: add docs
	ImageTypeStore               ImageType = "Store"
	ImageTypeFreezer             ImageType = "Freezer"
	ImageTypeFreezerCode         ImageType = "FreezerCode"
	ImageTypeProductDisplayClose ImageType = "ProductDisplayClose"
	ImageTypeProductDisplayOpen  ImageType = "ProductDisplayOpen"
	ImageTypeLowerFreezerPhoto   ImageType = "LowerFreezerPhoto"
	ImageTypeBackupPhoto         ImageType = "BackupPhoto"
	ImageTypeFreezerThermometer  ImageType = "FreezerThermometer"
)

func (s ImageType) String() string {
	return string(s)
}

// Freezer State Type.
type FreezerState string

const (
	FreezerStateClean    FreezerState = "Clean"
	FreezerStateNotClean FreezerState = "Not Clean"
	FreezerStateDirty    FreezerState = "Dirty"
)

func (s FreezerState) String() string {
	return string(s)
}

var SliceFreezerState = []string{
	string(FreezerStateClean),
	string(FreezerStateNotClean),
	string(FreezerStateDirty),
}

// End Freezer State Type.

// Freezer Positioning Type.
type FreezerPositioning string

const (
	FreezerPositioningOptimal        FreezerPositioning = "Optimal"
	FreezerPositioningAcceptable     FreezerPositioning = "Acceptable"
	FreezerPositioningNeedAdjustment FreezerPositioning = "Need Adjustment"
)

func (s FreezerPositioning) String() string {
	return string(s)
}

var SliceFreezerPositioning = []string{
	string(FreezerPositioningOptimal),
	string(FreezerPositioningAcceptable),
	string(FreezerPositioningNeedAdjustment),
}

// End Freezer Positioning Type.

// Freezer Capacity Status Type.
type FreezerCapacityStatus string

const (
	FreezerCapacityStatus_0  FreezerCapacityStatus = "0-20"
	FreezerCapacityStatus_20 FreezerCapacityStatus = "20-40"
	FreezerCapacityStatus_40 FreezerCapacityStatus = "40-60"
	FreezerCapacityStatus_60 FreezerCapacityStatus = "60-80"
	FreezerCapacityStatus_80 FreezerCapacityStatus = "80-100"
)

func (s FreezerCapacityStatus) String() string {
	return string(s)
}

var SliceFreezerCapacityStatus = []string{
	string(FreezerCapacityStatus_0),
	string(FreezerCapacityStatus_20),
	string(FreezerCapacityStatus_40),
	string(FreezerCapacityStatus_60),
	string(FreezerCapacityStatus_80),
}

// End Freezer Capacity Status Type.

// Freezer Thermometer State.
type FreezerThermometer string

const (
	FreezerThermometerGood FreezerThermometer = "Good"
	FreezerThermometerBad  FreezerThermometer = "Bad"
)

func (s FreezerThermometer) String() string {
	return string(s)
}

var SliceFreezerThermometer = []string{
	string(FreezerThermometerGood),
	string(FreezerThermometerBad),
}

// End Freezer Thermometer State.

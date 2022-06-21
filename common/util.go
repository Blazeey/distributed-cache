package common

func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// TODO: Not used - Remove if unnecessary
func ConvertUInt32ToInt64(values []uint32) []int64 {
	var convertedValues []int64 = make([]int64, 0)
	for _, v := range values {
		convertedValues = append(convertedValues, int64(v))
	}
	return convertedValues
}

func ConvertInt64ToUInt32(values []int64) []uint32 {
	var convertedValues []uint32 = make([]uint32, 0)
	for _, v := range values {
		convertedValues = append(convertedValues, uint32(v))
	}
	return convertedValues
}

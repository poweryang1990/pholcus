package spider_lib

// 基础包
import (
	"math"
)
type HouseSourceSetting struct{
    MaxPage int
    CityCode string
    Areas []string
}

func Round(f float64, places int) (float64) {
    shift := math.Pow(10, float64(places))
    return math.Floor(f * shift + .5) / shift
}
package shape_msgs

import (
	"github.com/aler9/goroslib/msg"
)

type Plane struct {
	msg.Package `ros:"shape_msgs"`
	Coef        [4]float64
}

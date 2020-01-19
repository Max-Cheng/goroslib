// Autogenerated with msg-import, do not edit.
package sensor_msgs

import (
	"github.com/aler9/goroslib/msg"
	"github.com/aler9/goroslib/msgs/std_msgs"
)

type MultiEchoLaserScan struct {
	Header         std_msgs.Header
	AngleMin       msg.Float32
	AngleMax       msg.Float32
	AngleIncrement msg.Float32
	TimeIncrement  msg.Float32
	ScanTime       msg.Float32
	RangeMin       msg.Float32
	RangeMax       msg.Float32
	Ranges         []LaserEcho
	Intensities    []LaserEcho
}
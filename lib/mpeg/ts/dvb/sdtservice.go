package dvb

import (
	"fmt"
	"strings"

	desc "github.com/puellanivis/breton/lib/mpeg/ts/descriptor"
)

type RunningStatus uint8

const (
	NotRunning RunningStatus = iota + 1
	Starting
	Pausing
	Running
	OffAir
)

var runningStatusNames = map[RunningStatus]string{
	NotRunning: "NOT_RUN",
	Starting:   "START",
	Pausing:    "PAUSE",
	Running:    "RUN",
	OffAir:     "OFF_AIR",
}

func (rs RunningStatus) String() string {
	if s, ok := runningStatusNames[rs]; ok {
		return s
	}

	return fmt.Sprintf("x%X", uint8(rs))
}

type SDTService struct {
	ServiceID     uint16
	EITSchedule   bool
	EITPresent    bool
	RunningStatus RunningStatus
	FreeCA        bool

	Descriptors []desc.Descriptor
}

func (s *SDTService) String() string {
	out := []string{
		"DVB:Service",
		fmt.Sprintf("ID:x%04x", s.ServiceID),
	}

	if s.EITSchedule {
		out = append(out, "EIT_SCHED")
	}

	if s.EITPresent {
		out = append(out, "EIT_PRES")
	}

	out = append(out, fmt.Sprint(s.RunningStatus))

	if s.FreeCA {
		out = append(out, "FreeCA")
	}

	for _, d := range s.Descriptors {
		out = append(out, fmt.Sprintf("Desc:%v", d))
	}

	return fmt.Sprintf("{%s}", strings.Join(out, " "))
}

const (
	flagEITSchedule = 0x02
	flagEITPresent  = 0x01

	shiftRunningStatus = 5
	maskRunningStatus  = 0x07

	flagFreeCA = 0x10
)

func (s *SDTService) Marshal() ([]byte, error) {
	b := make([]byte, 5)

	b[0] = byte((s.ServiceID >> 8) & 0xff)
	b[1] = byte(s.ServiceID & 0xff)
	b[2] = 0xFC

	if s.EITSchedule {
		b[2] |= flagEITSchedule
	}

	if s.EITPresent {
		b[2] |= flagEITPresent
	}

	b[3] = byte((s.RunningStatus & maskRunningStatus) << shiftRunningStatus)
	if s.FreeCA {
		b[3] |= flagFreeCA
	}

	loopLen := 0
	for _, d := range s.Descriptors {
		db, err := d.Marshal()
		if err != nil {
			return nil, err
		}

		loopLen += len(db)
		b = append(b, db...)
	}

	b[3] |= byte((loopLen >> 8) & 0x0F)
	b[4] = byte(loopLen & 0xFF)

	return b, nil
}

func (s *SDTService) Unmarshal(b []byte) (int, error) {
	s.ServiceID = uint16(b[0])<<8 | uint16(b[1])
	s.EITSchedule = b[2]&flagEITSchedule != 0
	s.EITPresent = b[2]&flagEITPresent != 0
	s.RunningStatus = RunningStatus((b[3] >> shiftRunningStatus) & maskRunningStatus)
	s.FreeCA = b[3]&flagFreeCA != 0

	loopLen := int(b[3]&0x0F)<<8 | int(b[4])

	start := 0
	b = b[5:]

	for start < loopLen {
		pd, err := desc.Unmarshal(b[start:])
		if err != nil {
			return start, err
		}

		start += pd.Len()
		s.Descriptors = append(s.Descriptors, pd)
	}

	return start + 5, nil
}

package egts

import (
	"bytes"
	"fmt"
)

type SrAbsDigSensData struct {
	SensorNumber uint16 `json:"DSN"`
	SensorState  uint8  `json:"DSST"`
}

func (s *SrAbsDigSensData) Decode(content []byte) error {
	if len(content) < 2 {
		return fmt.Errorf("неверная длина контента sr_abs_dig_sens_data: %d", len(content))
	}
	b0 := content[0]
	b1 := content[1]
	s.SensorState = b0 & 0x0F
	s.SensorNumber = uint16(b1)<<4 | uint16(b0>>4)
	return nil
}

func (s *SrAbsDigSensData) Encode() ([]byte, error) {
	if s.SensorNumber > 0x0FFF {
		return nil, fmt.Errorf("некорректный номер цифрового входа: %d", s.SensorNumber)
	}
	b0 := uint8((s.SensorNumber&0x0F)<<4) | (s.SensorState & 0x0F)
	b1 := uint8(s.SensorNumber >> 4)
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(b0)
	buf.WriteByte(b1)
	return buf.Bytes(), nil
}

func (s *SrAbsDigSensData) Length() uint16 {
	return 2
}

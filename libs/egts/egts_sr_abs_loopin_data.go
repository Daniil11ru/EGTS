package egts

import (
	"fmt"
)

type SrAbsLoopinData struct {
	LoopInNumber uint16 `json:"LIN"`
	LoopInState  uint8  `json:"LIS"`
}

func (s *SrAbsLoopinData) Decode(content []byte) error {
	if len(content) < 2 {
		return fmt.Errorf("неверная длина контента sr_abs_loopin_data: %d", len(content))
	}
	b0 := content[0]
	b1 := content[1]
	s.LoopInState = b0 & 0x0F
	s.LoopInNumber = uint16(b0>>4)&0x0F | uint16(b1)<<4
	return nil
}

func (s *SrAbsLoopinData) Encode() ([]byte, error) {
	if s.LoopInState > 0x0F {
		return nil, fmt.Errorf("некорректное значение LIS: %d", s.LoopInState)
	}
	if s.LoopInNumber > 0x0FFF {
		return nil, fmt.Errorf("некорректный номер LIN: %d", s.LoopInNumber)
	}
	b0 := uint8((s.LoopInNumber&0x0F)<<4) | (s.LoopInState & 0x0F)
	b1 := uint8(s.LoopInNumber >> 4)
	return []byte{b0, b1}, nil
}

func (s *SrAbsLoopinData) Length() uint16 {
	return 2
}

package egts

import (
	"bytes"
	"fmt"
	"strconv"
)

type SrLoopinData struct {
	LoopInFieldExists1 string `json:"LIFE1"`
	LoopInFieldExists2 string `json:"LIFE2"`
	LoopInFieldExists3 string `json:"LIFE3"`
	LoopInFieldExists4 string `json:"LIFE4"`
	LoopInFieldExists5 string `json:"LIFE5"`
	LoopInFieldExists6 string `json:"LIFE6"`
	LoopInFieldExists7 string `json:"LIFE7"`
	LoopInFieldExists8 string `json:"LIFE8"`
	LoopInState1       uint8  `json:"LIS1"`
	LoopInState2       uint8  `json:"LIS2"`
	LoopInState3       uint8  `json:"LIS3"`
	LoopInState4       uint8  `json:"LIS4"`
	LoopInState5       uint8  `json:"LIS5"`
	LoopInState6       uint8  `json:"LIS6"`
	LoopInState7       uint8  `json:"LIS7"`
	LoopInState8       uint8  `json:"LIS8"`
}

func (l *SrLoopinData) Decode(content []byte) error {
	buf := bytes.NewReader(content)
	flags, err := buf.ReadByte()
	if err != nil {
		return fmt.Errorf("не удалось получить байт флагов sr_loopin_data: %v", err)
	}
	bits := fmt.Sprintf("%08b", flags)
	l.LoopInFieldExists8 = bits[:1]
	l.LoopInFieldExists7 = bits[1:2]
	l.LoopInFieldExists6 = bits[2:3]
	l.LoopInFieldExists5 = bits[3:4]
	l.LoopInFieldExists4 = bits[4:5]
	l.LoopInFieldExists3 = bits[5:6]
	l.LoopInFieldExists2 = bits[6:7]
	l.LoopInFieldExists1 = bits[7:]

	readState := func() (uint8, error) {
		b, e := buf.ReadByte()
		if e != nil {
			return 0, e
		}
		return b, nil
	}

	if l.LoopInFieldExists1 == "1" {
		if l.LoopInState1, err = readState(); err != nil {
			return fmt.Errorf("не удалось получить LIS1: %v", err)
		}
	}
	if l.LoopInFieldExists2 == "1" {
		if l.LoopInState2, err = readState(); err != nil {
			return fmt.Errorf("не удалось получить LIS2: %v", err)
		}
	}
	if l.LoopInFieldExists3 == "1" {
		if l.LoopInState3, err = readState(); err != nil {
			return fmt.Errorf("не удалось получить LIS3: %v", err)
		}
	}
	if l.LoopInFieldExists4 == "1" {
		if l.LoopInState4, err = readState(); err != nil {
			return fmt.Errorf("не удалось получить LIS4: %v", err)
		}
	}
	if l.LoopInFieldExists5 == "1" {
		if l.LoopInState5, err = readState(); err != nil {
			return fmt.Errorf("не удалось получить LIS5: %v", err)
		}
	}
	if l.LoopInFieldExists6 == "1" {
		if l.LoopInState6, err = readState(); err != nil {
			return fmt.Errorf("не удалось получить LIS6: %v", err)
		}
	}
	if l.LoopInFieldExists7 == "1" {
		if l.LoopInState7, err = readState(); err != nil {
			return fmt.Errorf("не удалось получить LIS7: %v", err)
		}
	}
	if l.LoopInFieldExists8 == "1" {
		if l.LoopInState8, err = readState(); err != nil {
			return fmt.Errorf("не удалось получить LIS8: %v", err)
		}
	}
	return nil
}

func (l *SrLoopinData) Encode() ([]byte, error) {
	var (
		buf   = new(bytes.Buffer)
		err   error
		flags uint64
	)

	bits := l.LoopInFieldExists8 +
		l.LoopInFieldExists7 +
		l.LoopInFieldExists6 +
		l.LoopInFieldExists5 +
		l.LoopInFieldExists4 +
		l.LoopInFieldExists3 +
		l.LoopInFieldExists2 +
		l.LoopInFieldExists1

	if flags, err = strconv.ParseUint(bits, 2, 8); err != nil {
		return nil, fmt.Errorf("не удалось сформировать байт флагов sr_loopin_data: %v", err)
	}
	if err = buf.WriteByte(uint8(flags)); err != nil {
		return nil, fmt.Errorf("не удалось записать байт флагов sr_loopin_data: %v", err)
	}

	writeState := func(exists string, val uint8) error {
		if exists == "1" {
			if err := buf.WriteByte(val); err != nil {
				return err
			}
		}
		return nil
	}

	if err = writeState(l.LoopInFieldExists1, l.LoopInState1); err != nil {
		return nil, fmt.Errorf("не удалось записать LIS1: %v", err)
	}
	if err = writeState(l.LoopInFieldExists2, l.LoopInState2); err != nil {
		return nil, fmt.Errorf("не удалось записать LIS2: %v", err)
	}
	if err = writeState(l.LoopInFieldExists3, l.LoopInState3); err != nil {
		return nil, fmt.Errorf("не удалось записать LIS3: %v", err)
	}
	if err = writeState(l.LoopInFieldExists4, l.LoopInState4); err != nil {
		return nil, fmt.Errorf("не удалось записать LIS4: %v", err)
	}
	if err = writeState(l.LoopInFieldExists5, l.LoopInState5); err != nil {
		return nil, fmt.Errorf("не удалось записать LIS5: %v", err)
	}
	if err = writeState(l.LoopInFieldExists6, l.LoopInState6); err != nil {
		return nil, fmt.Errorf("не удалось записать LIS6: %v", err)
	}
	if err = writeState(l.LoopInFieldExists7, l.LoopInState7); err != nil {
		return nil, fmt.Errorf("не удалось записать LIS7: %v", err)
	}
	if err = writeState(l.LoopInFieldExists8, l.LoopInState8); err != nil {
		return nil, fmt.Errorf("не удалось записать LIS8: %v", err)
	}

	return buf.Bytes(), nil
}

func (l *SrLoopinData) Length() uint16 {
	b, err := l.Encode()
	if err != nil {
		return 0
	}
	return uint16(len(b))
}

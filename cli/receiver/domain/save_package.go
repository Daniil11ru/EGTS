package domain

import (
	"math/bits"

	aux "github.com/daniil11ru/egts/cli/receiver/repository/auxiliary"
	"github.com/daniil11ru/egts/cli/receiver/repository/movement"
	packet "github.com/daniil11ru/egts/cli/receiver/repository/movement/util"
	source "github.com/daniil11ru/egts/cli/receiver/source/auxiliary"
	log "github.com/sirupsen/logrus"
)

type SavePackage struct {
	VehicleMovementRepository      *movement.VehicleMovementRepository
	AuxiliaryInformationRepository aux.AuxiliaryInformationRepository

	IDToVehicleDirectory map[int]source.VehicleDirectory
	OIDToVehicleID       map[int]int
}

func (domain *SavePackage) Initialize() {
	directories, _ := domain.AuxiliaryInformationRepository.GetAllDirectories()

	for _, d := range directories {
		domain.IDToVehicleDirectory[d.ID] = d
	}
}

func isPrefixBytes(a, b uint64, n int) bool {
	width := (bits.Len64(b) + 7) / 8
	shift := uint((width - n) * 8)
	return a == b>>shift
}

func isSuffixBytes(a, b uint64, n int) bool {
	width := (bits.Len64(b) + 7) / 8
	if n >= width {
		return a == b
	}
	return a == b&((1<<(uint(n)*8))-1)
}

func isPrefixDigits(a, b uint64, n int) bool {
	digits := 1
	for tmp := b; tmp >= 10; tmp /= 10 {
		digits++
	}
	div := uint64(1)
	for i := 0; i < digits-n; i++ {
		div *= 10
	}
	return a == b/div
}

func isSuffixDigits(a, b uint64, n int) bool {
	mod := uint64(1)
	for i := 0; i < n; i++ {
		mod *= 10
	}
	return a == b%mod
}

func digitCount(n uint64) int {
	if n == 0 {
		return 1
	}
	count := 0
	for n > 0 {
		count++
		n /= 10
	}
	return count
}

func byteCount(n uint64) int {
	if n == 0 {
		return 1
	}
	return (bits.Len64(n) + 7) / 8
}

func (domain *SavePackage) getVehicleIDFromOID(OID uint32, vehicles []source.Vehicle) int {
	for _, v := range vehicles {
		directory := domain.IDToVehicleDirectory[v.DirectoryID]
		IMEI := v.IMEI

		var segmentLength int16

		if directory.SegmentLength.Valid {
			segmentLength = directory.SegmentLength.Int16
		} else {
			segmentLength = 0
		}

		var unitCount uint8

		if segmentLength != 0 {
			unitCount = uint8(segmentLength)
		} else {
			if directory.ExtractionUnitType == source.ExtractionUnitTypeBytes {
				unitCount = uint8(byteCount(uint64(OID)))
			} else if directory.ExtractionUnitType == source.ExtractionUnitTypeDigits {
				unitCount = uint8(digitCount(uint64(OID)))
			}
		}

		if directory.ExtractionUnitType == source.ExtractionUnitTypeBytes && directory.ExtractionPositionType == source.ExtractionPositionTypePrefix {
			if isPrefixBytes(uint64(OID), uint64(IMEI), int(unitCount)) {
				return v.ID
			}
		} else if directory.ExtractionUnitType == source.ExtractionUnitTypeBytes && directory.ExtractionPositionType == source.ExtractionPositionTypeSuffix {
			if isSuffixBytes(uint64(OID), uint64(IMEI), int(unitCount)) {
				return v.ID
			}
		} else if directory.ExtractionUnitType == source.ExtractionUnitTypeDigits && directory.ExtractionPositionType == source.ExtractionPositionTypePrefix {
			if isPrefixDigits(uint64(OID), uint64(IMEI), int(unitCount)) {
				return v.ID
			}
		} else if directory.ExtractionUnitType == source.ExtractionUnitTypeDigits && directory.ExtractionPositionType == source.ExtractionPositionTypeSuffix {
			if isSuffixDigits(uint64(OID), uint64(IMEI), int(unitCount)) {
				return v.ID
			}
		}
	}

	return -1
}

func (domain *SavePackage) Run(data *packet.NavRecord, providerIP string) {
	OID := data.Client
	vehicleID, OK := domain.OIDToVehicleID[int(OID)]
	if !OK {
		vehicles, _ := domain.AuxiliaryInformationRepository.GetVehiclesByProviderIP(providerIP)
		vehicleID = domain.getVehicleIDFromOID(OID, vehicles)
	}

	if vehicleID >= 0 {
		domain.VehicleMovementRepository.Save(data, vehicleID)
		domain.OIDToVehicleID[int(OID)] = vehicleID
	} else {
		log.Warnf("Не удалось найти машину по OID %d, телематические данные не были записаны", OID)
	}
}

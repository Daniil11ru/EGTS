package out

type Track struct {
	VehicleId int32   `json:"vehicle_id"`
	Points    []Point `json:"points"`
}

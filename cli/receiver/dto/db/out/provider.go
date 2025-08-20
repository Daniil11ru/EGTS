package out

type Provider struct {
	ID   int32  `json:"id" gorm:"column:id"`
	Name string `json:"name"`
}

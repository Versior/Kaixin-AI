package model

// CanvasProject stores a user's inspiration canvas JSON so it can follow the account across browsers.
type CanvasProject struct {
	ID        string `json:"id" gorm:"primaryKey"`
	UserID    string `json:"userId" gorm:"primaryKey;index"`
	User      User   `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Title     string `json:"title"`
	Payload   string `json:"payload" gorm:"type:text"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

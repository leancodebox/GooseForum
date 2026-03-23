package articleCategory

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

const tableName = "article_category"

// pid
const pid = "id"

// fieldCategory
const fieldCategory = "category"

// fieldCreatedAt
const fieldCreatedAt = "created_at"

// fieldUpdatedAt
const fieldUpdatedAt = "updated_at"

// fieldIcon
const fieldIcon = "icon"

// fieldColor
const fieldColor = "color"

// fieldSlug
const fieldSlug = "slug"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`                //
	Category  string    `gorm:"column:category;type:varchar(64);not null;default:'';" json:"category"` //
	Desc      string    `gorm:"column:desc;type:varchar(255);not null;default:'';" json:"desc"`        //
	Icon      string    `gorm:"column:icon;type:varchar(255);not null;default:'';" json:"icon"`        //
	Color     string    `gorm:"column:color;type:varchar(255);not null;default:'';" json:"color"`      //
	Slug      string    `gorm:"column:slug;type:varchar(255);not null;default:'';" json:"slug"`        //
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`          //
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

// func (itself *Entity) BeforeSave(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeCreate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterCreate(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeUpdate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterUpdate(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterSave(tx *gorm.DB) (err error) {}
// func (itself *Entity) BeforeDelete(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterDelete(tx *gorm.DB) (err error) {}
// func (itself *Entity) AfterFind(tx *gorm.DB) (err error) {}

func (itself *Entity) TableName() string {
	return tableName
}

func (itself *Entity) AfterFind(tx *gorm.DB) (err error) {
	if itself.Color == "" {
		itself.Color = itself.CalculateColor()
	}
	return
}

func (itself *Entity) BeforeSave(tx *gorm.DB) (err error) {
	if itself.Color == "" {
		itself.Color = itself.CalculateColor()
	}
	return
}

// CalculateColor 根据分类名称计算一个固定的 HSL 颜色，返回 Hex 格式
func (itself *Entity) CalculateColor() string {
	var hash uint32 = 0
	for i := 0; i < len(itself.Category); i++ {
		hash = uint32(itself.Category[i]) + ((hash << 5) - hash)
	}

	h := int(hash % 360)
	s := 65
	l := 55

	return hslToHex(h, s, l)
}

func hslToHex(h, s, l int) string {
	s1 := float64(s) / 100
	l1 := float64(l) / 100

	c := (1 - MathAbs(2*l1-1)) * s1
	x := c * (1 - MathAbs(math.Mod(float64(h)/60, 2)-1))
	m := l1 - c/2

	var r, g, b float64
	if h < 60 {
		r, g, b = c, x, 0
	} else if h < 120 {
		r, g, b = x, c, 0
	} else if h < 180 {
		r, g, b = 0, c, x
	} else if h < 240 {
		r, g, b = 0, x, c
	} else if h < 300 {
		r, g, b = x, 0, c
	} else {
		r, g, b = c, 0, x
	}

	ri := int((r + m) * 255)
	gi := int((g + m) * 255)
	bi := int((b + m) * 255)

	return fmt.Sprintf("#%02x%02x%02x", ri, gi, bi)
}

func MathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

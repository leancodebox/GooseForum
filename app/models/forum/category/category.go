package category

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

const tableName = "category"

type Entity struct {
	Id        uint64    `gorm:"primaryKey;column:id;autoIncrement;not null;" json:"id"`
	Name      string    `gorm:"column:name;type:varchar(64);not null;default:'';" json:"name"`
	Desc      string    `gorm:"column:desc;type:varchar(255);not null;default:'';" json:"desc"`
	Icon      string    `gorm:"column:icon;type:varchar(255);not null;default:'';" json:"icon"`
	Color     string    `gorm:"column:color;type:varchar(255);not null;default:'';" json:"color"`
	Slug      string    `gorm:"column:slug;type:varchar(255);not null;default:'';" json:"slug"`
	Sort      int       `gorm:"column:sort;type:int;not null;default:0;index:idx_category_sort,priority:1;" json:"sort"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;<-:create;" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;" json:"updatedAt"`
}

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

func (itself *Entity) CalculateColor() string {
	var hash uint32
	for i := 0; i < len(itself.Name); i++ {
		hash = uint32(itself.Name[i]) + ((hash << 5) - hash)
	}

	h := int(hash % 360)
	s := 65
	l := 55

	return hslToHex(h, s, l)
}

func hslToHex(h, s, l int) string {
	s1 := float64(s) / 100
	l1 := float64(l) / 100

	c := (1 - mathAbs(2*l1-1)) * s1
	x := c * (1 - mathAbs(math.Mod(float64(h)/60, 2)-1))
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

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

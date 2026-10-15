package trusthistory

import "github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"

func Create(h *History) error { return dbconnect.Connect().Create(h).Error }
func List(userID, after uint64) ([]History, error) {
	rows := []History{}
	q := dbconnect.Connect().Where("user_id = ?", userID)
	if after > 0 {
		q = q.Where("id < ?", after)
	}
	err := q.Order("id DESC").Limit(30).Find(&rows).Error
	return rows, err
}

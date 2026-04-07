package domain

import "errors"

var (
	ErrInvalidWorkday = errors.New("workday must be between 1 and 7")
)

type Workday int

const (
	Monday    Workday = 1 << iota // 1
	Tuesday                       // 2
	Wednesday                     // 4
	Thursday                      // 8
	Friday                        // 16
	Saturday                      // 32
	Sunday                        // 64
)

func NewWorkday(days []int) (Workday, error) {
	w := Workday(0)
	for _, d := range days {
		if d < 1 || d > 7 {
			return 0, ErrInvalidWorkday
		}
		w |= 1 << (d - 1)
	}
	return w, nil
}

func (w Workday) IsWorkday(d Workday) bool {
	return w&d != 0
}

func (w Workday) ToSlice() []int {
	days := make([]int, 0, 7)
	for i := 0; i < 7; i++ {
		if w&(1<<i) != 0 {
			days = append(days, i+1)
		}
	}
	return days
}

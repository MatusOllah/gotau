package observance

import "time"

// IsMikuDay checks if the given time is Miku Day (March 9th).
func IsMikuDay(t time.Time) bool {
	return t.Month() == time.March && t.Day() == 9
}

// IsMikuBirthday checks if the given time is Hatsune Miku's birthday (August 31st).
func IsMikuBirthday(t time.Time) bool {
	return t.Month() == time.August && t.Day() == 31
}

// IsTetoBirthday checks if the given time is Kasane Teto's birthday (April 1st i.e. April Fools' Day).
func IsTetoBirthday(t time.Time) bool {
	return t.Month() == time.April && t.Day() == 1
}

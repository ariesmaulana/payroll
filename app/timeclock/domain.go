package timeclock

import "time"

func countWorkdays(start, end time.Time) int {
	start = start.Truncate(24 * time.Hour)
	end = end.Truncate(24 * time.Hour)

	count := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		weekday := d.Weekday()
		if weekday >= time.Monday && weekday <= time.Friday {
			count++
		}
	}
	return count
}

func isFullMonthAttendance(start, end time.Time, totalAttendances int) bool {
	workingDays := countWorkdays(start, end)
	return totalAttendances == workingDays
}

func calculateProratedSalary(
	baseSalaries map[int]int, // key is userId and value is baseSalary
	attendances map[int]int, // key is userId and value is totalAttendance
	periodStart time.Time,
	periodEnd time.Time,
) map[int]int {
	// key is userId and value is totalSalary
	result := make(map[int]int)

	workdays := countWorkdays(periodStart, periodEnd)
	if workdays == 0 {
		return result
	}

	for userID, baseSalary := range baseSalaries {
		attendance := attendances[userID]
		salary := baseSalary * attendance / workdays
		result[userID] = salary
	}

	return result
}

func sumValuesMap(params map[int]int) int {
	total := 0
	for _, m := range params {
		total += m
	}
	return total
}

func calculateOvertimeSalary(
	baseSalaryMap map[int]int, // key is userId and value is baseSalary
	attendanceMap map[int]int, // key is userId and value is totalAttendance
	overtimeMap map[int]int, // key is userId and value is totalOverTimeHoures
	periodStart time.Time,
	periodEnd time.Time,
) map[int]int {
	// key is userId and value is total overtimePay
	result := make(map[int]int)

	workdays := countWorkdays(periodStart, periodEnd)
	if workdays == 0 {
		return result
	}

	for userID, baseSalary := range baseSalaryMap {

		attendanceDays := attendanceMap[userID]
		overtimeHours := overtimeMap[userID]
		if attendanceDays == 0 || overtimeHours == 0 || workdays == 0 {
			continue
		}

		proratedSalary := baseSalary * attendanceDays / workdays

		perDay := proratedSalary / attendanceDays
		perHour := perDay / 8 // assums working hour is 8

		overtimePay := overtimeHours * perHour * 2

		result[userID] = overtimePay
	}

	return result
}

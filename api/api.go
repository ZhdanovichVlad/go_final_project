package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

//func NextDateCopy(now time.Time, date string, repeat string) (string, error) {
//	if repeat == "" {
//		return "", nil
//	}
//	dateTask, err := time.Parse("20060102", date)
//	if err != nil {
//		return "", fmt.Errorf("failed to convert string to date", err)
//	}
//	codeRepeat := strings.Split(repeat, " ")
//	if len(codeRepeat) != 2 && !strings.EqualFold(codeRepeat[0], "y") {
//		return "", fmt.Errorf("string conversion error repeat")
//	}
//	switch strings.ToLower(codeRepeat[0]) {
//	case "y":
//		dateTask = addDateTask(now, dateTask, 1, 0, 0)
//	case "d":
//		digit, err := convetToIntAncCheck(codeRepeat[1])
//		if err != nil {
//			return "", err
//		}
//		dateTask = addDateTask(now, dateTask, 0, 0, digit)
//	case "w":
//
//		dayOfWeek
//		// Текущая дата и время
//		now := time.Now()
//
//		// Вычисление разницы в днях
//		daysToTarget := int(time.Weekday(dayOfWeek-1) - now.Weekday()) // -1 для корректного сопоставления с time.Weekday
//		if daysToTarget < 0 {
//			daysToTarget += 7
//		}
//
//		// Добавление дней к текущей дате
//		targetDate := now.AddDate(0, 0, daysToTarget)
//
//	default:
//		return "", fmt.Errorf("unknown key in repeat")
//	}
//	return dateTask.Format("20060102"), nil
//}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", nil
	}
	dateTask, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("failed to convert string to date", err)
	}
	repeatRune := []rune(repeat)
	firstRepeatLetter := string(repeatRune[0])
	if firstRepeatLetter != "y" && firstRepeatLetter != "d" && firstRepeatLetter != "w" && firstRepeatLetter != "m" {
		return "", fmt.Errorf("wrong repetition rule")
	}
	var dateAnswear time.Time
	switch firstRepeatLetter {
	case "y":
		dateAnswear = addDateTask(now, dateTask, 1, 0, 0)
	case "d":
		repeatSplit := strings.Split(repeat, " ")
		if len(repeatSplit) != 2 {
			return "", fmt.Errorf("wrong repetition rule")
		}
		digit, err := convetToIntAncCheck(repeatSplit[1])
		if err != nil {
			return "", err
		}
		dateAnswear = addDateTask(now, dateTask, 0, 0, digit)
	case "w":
		repeatCheak := strings.Split(repeat, " ")
		if len(repeatCheak) != 2 {
			return "", fmt.Errorf("wrong repetition rule")
		}
		lastRepeatRune := repeatRune[2:]
		repeatSplit := strings.Split(string(lastRepeatRune), ",")
		dateAnswear = now.AddDate(1, 0, 0)
		dateComparison := dateTask
		for _, val := range repeatSplit {
			valInt, err := strconv.Atoi(val)
			if err != nil {
				return "", fmt.Errorf("wrong repetition rule")
			}
			if 1 <= valInt && valInt <= 7 {
				weekday := time.Weekday((valInt + 7) % 7)
				daysToTarget := int(weekday - dateTask.Weekday())
				if daysToTarget < 0 {
					daysToTarget += 7
				}
				dateComparison = dateComparison.AddDate(0, 0, daysToTarget)
				if now.After(dateComparison) {
					dateComparison = addDateTask(now, dateComparison, 0, 0, 7)
				}
				if dateAnswear.After(dateComparison) {
					dateAnswear = dateComparison
				}
				fmt.Println("dateAnswear в цикле", dateAnswear)
			} else {
				return "", fmt.Errorf("error when entering the day of the week")
			}
		}
	case "m":
		lastRepeatRune := repeatRune[2:]
		lastStringSplit := strings.Split(string(lastRepeatRune), " ")
		var masDays []string
		var masMounth []string
		var flagWithMounth bool
		dateAnswear = time.Date(now.Year()+1, now.Month(), 1, 0, 0, 0, 0, now.Location())
		dateComparison := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		if len(lastStringSplit) == 2 {
			masDays = strings.Split(lastStringSplit[0], ",")
			masMounth = strings.Split(lastStringSplit[1], ",")
			flagWithMounth = true
		}
		if flagWithMounth { // если на входе в даты и месяца
			for _, valDays := range masDays {
				valDaysInt, err := strconv.Atoi(valDays)
				if err != nil {
					return "", fmt.Errorf("wrong repetition rule")
				}
				if 1 < valDaysInt && valDaysInt > 31 {
					return "", fmt.Errorf("wrong repetition rule")
				}
				for _, valMounth := range masMounth {
					valMounthInt, err := strconv.Atoi(valMounth)
					if err != nil {
						return "", fmt.Errorf("wrong repetition rule")
					}
					if 1 < valMounthInt && valMounthInt > 12 {
						return "", fmt.Errorf("wrong repetition rule")
					}
					if valDaysInt == -1 && valDaysInt == -2 {
						dateComparison = time.Date(now.Year(), time.Month(valMounthInt)+1, valDaysInt, 0, 0, 0, 0, now.Location())
					} else {
						dateComparison = time.Date(now.Year(), time.Month(valMounthInt), valDaysInt, 0, 0, 0, 0, now.Location())
					}
					fmt.Println("значениея ", valDays, " ", valMounth)
					fmt.Println("dateComparison ", dateComparison)
					fmt.Println("dateAnswear    ", dateAnswear)
					fmt.Println("now.Before(dateComparison) ", now.Before(dateComparison))
					fmt.Println("dateAnswear.Before(dateComparison)", dateAnswear.Before(dateComparison))
					fmt.Println("result", now.Before(dateComparison) && dateAnswear.Before(dateComparison))
					if now.Before(dateComparison) && dateAnswear.After(dateComparison) && !now.Equal(dateComparison) {
						dateAnswear = dateComparison
					}
				}
			}

		} else { /// если на входе в повторение только даты
			repeatSplit := strings.Split(lastStringSplit[0], ",")
			for _, val := range repeatSplit {
				valInt, err := strconv.Atoi(val)
				if err != nil {
					return "", fmt.Errorf("wrong repetition rule")
				}
				if valInt == -1 || valInt == -2 {
					dateComparison = time.Date(dateTask.Year(), dateTask.Month()+1, valInt+1, 0, 0, 0, 0, dateTask.Location())
					fmt.Println("значение valInt", valInt)
					fmt.Println("время сегодня", now)
					fmt.Println("полученное время по условиям", dateComparison)
					if dateAnswear.After(dateComparison) {
						fmt.Println("перезапись времени")
						dateAnswear = dateComparison
					}
				} else if 1 <= valInt && valInt < 31 {
					dateComparison = time.Date(dateTask.Year(), dateTask.Month(), valInt, 0, 0, 0, 0, dateTask.Location())
					if now.After(dateComparison) || now.Equal(dateComparison) {
						dateComparison = addDateTask(now, dateComparison, 0, 1, 0)
					}
					if dateAnswear.After(dateComparison) {
						dateAnswear = dateComparison
					}

				} else if valInt == 31 {
					nowMonth := dateTask.Month()
					if nowMonth == time.February || nowMonth == time.April || nowMonth == time.June || nowMonth == time.September || nowMonth == time.November {
						nowMonth = nowMonth + 1
					}
					dateComparison = time.Date(dateTask.Year(), nowMonth, valInt, 0, 0, 0, 0, dateTask.Location())
					if now.After(dateComparison) || now.Equal(dateComparison) {
						dateComparison = addDateTask(now, dateComparison, 0, 1, 0)
					}
					if dateAnswear.After(dateComparison) {
						dateAnswear = dateComparison
					}

				} else {
					return "", fmt.Errorf("wrong repetition rule")
				}
			}
			fmt.Println("вывод ответа dateAnswear", dateAnswear)
		}

	}
	return dateAnswear.Format("20060102"), nil
}

func convetToIntAncCheck(num string) (int, error) {
	digit, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}
	if digit >= 400 || digit < 0 {
		return 0, fmt.Errorf("the number of days to reschedule the task is greater than 400 or a negative number is specified")
	}
	return digit, nil
}

func addDateTask(now time.Time, dateTask time.Time, y int, m int, d int) time.Time {
	dateTask = dateTask.AddDate(y, m, d)

	for dateTask.Before(now) {
		dateTask = dateTask.AddDate(y, m, d)
	}
	return dateTask
}

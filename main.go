package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type ScheduleParams struct {
	LpuID                       int
	DoctorID                    int
	ScheduleStartDate           string
	ScheduleEndDate             string
	ScheduleWorkDays            []time.Weekday
	ScheduleWeekParity          string
	ScheduleDayParity           string
	ScheduleStartTime           string
	ScheduleEndTime             string
	ScheduleAppointmentDuration int
	ScheduleOfficeNumber        int
}

type ScheduleCell struct {
	LpuID                 int
	DoctorID              int
	ScheduleCellDate      string
	ScheduleCellTime      string
	ScheduleStatus        string
	ScheduleReceptionType string
	ScheduleOfficeNumber  int
	ScheduleComment       string
}

func main() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5500"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/schedule", createSchedule)
	router.GET("/schedule/:doctor_id", getSchedule)
	router.Run(":8080")
}

func createSchedule(c *gin.Context) {
	var params ScheduleParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cells := generateScheduleCells(params)
	if err := insertScheduleCells(cells); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при вставке данных: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Расписание успешно создано"})
}

func getSchedule(c *gin.Context) {
	doctorID := c.Param("doctor_id")
	db, err := sql.Open("mysql", "root:root!!!!@/medical_schedule")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка подключения к базе данных: " + err.Error()})
		return
	}
	defer db.Close()

	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 14)
	query := `SELECT * FROM schedule WHERE doctor_id = ? AND schedule_cell_date BETWEEN ? AND ? ORDER BY schedule_cell_date, schedule_cell_time`
	rows, err := db.Query(query, doctorID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка запроса к базе данных: " + err.Error()})
		return
	}
	defer rows.Close()

	var cells []ScheduleCell
	for rows.Next() {
		var cell ScheduleCell
		if err := rows.Scan(&cell.LpuID, &cell.DoctorID, &cell.ScheduleCellDate, &cell.ScheduleCellTime, &cell.ScheduleStatus, &cell.ScheduleReceptionType, &cell.ScheduleOfficeNumber, &cell.ScheduleComment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения данных: " + err.Error()})
			return
		}
		cells = append(cells, cell)
	}
	if len(cells) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Нет доступных ячеек расписания для данного врача"})
		return
	}
	c.JSON(http.StatusOK, cells)
}

func insertScheduleCells(cells []ScheduleCell) error {
	db, err := sql.Open("mysql", "root:root!!!!@/medical_schedule")
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare(`
        INSERT INTO schedule (lpu_id, doctor_id, schedule_cell_date, schedule_cell_time, schedule_status, schedule_reception_type, schedule_office_number, schedule_comment)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, cell := range cells {
		_, err := stmt.Exec(
			cell.LpuID,
			cell.DoctorID,
			cell.ScheduleCellDate,
			cell.ScheduleCellTime,
			cell.ScheduleStatus,
			cell.ScheduleReceptionType,
			cell.ScheduleOfficeNumber,
			cell.ScheduleComment,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func parseTime(t string) time.Time {
	parsedTime, _ := time.Parse("15:04", t)
	return parsedTime
}

func generateTimeSlots(start, end time.Time, duration int) []string {
	var slots []string
	for t := start; t.Before(end); t = t.Add(time.Minute * time.Duration(duration)) {
		slots = append(slots, t.Format("15:04"))
	}
	return slots
}

func generateScheduleCells(params ScheduleParams) []ScheduleCell {
	var cells []ScheduleCell
	startDate, _ := time.Parse("2006-01-02", params.ScheduleStartDate)
	endDate, _ := time.Parse("2006-01-02", params.ScheduleEndDate)
	startTime := parseTime(params.ScheduleStartTime)
	endTime := parseTime(params.ScheduleEndTime)

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		weekday := d.Weekday()
		dayOfMonth := d.Day()
		_, weekNumber := d.ISOWeek()

		isWorkDay := false
		for _, wd := range params.ScheduleWorkDays {
			if weekday == wd {
				isWorkDay = true
				break
			}
		}

		if !isWorkDay {
			continue
		}

		if params.ScheduleWeekParity == "четная" && weekNumber%2 != 0 {
			continue
		}
		if params.ScheduleWeekParity == "нечетная" && weekNumber%2 == 0 {
			continue
		}

		if (params.ScheduleDayParity == "четный" && dayOfMonth%2 != 0) || (params.ScheduleDayParity == "нечетный" && dayOfMonth%2 == 0) {
			continue
		}

		slots := generateTimeSlots(startTime, endTime, params.ScheduleAppointmentDuration)
		for _, slot := range slots {
			cell := ScheduleCell{
				LpuID:                 params.LpuID,
				DoctorID:              params.DoctorID,
				ScheduleCellDate:      d.Format("2006-01-02"),
				ScheduleCellTime:      slot,
				ScheduleStatus:        "доступно",
				ScheduleReceptionType: "первичный",
				ScheduleOfficeNumber:  params.ScheduleOfficeNumber,
				ScheduleComment:       "",
			}
			cells = append(cells, cell)
		}
	}

	return cells
}

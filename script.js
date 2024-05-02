document.getElementById('scheduleForm').addEventListener('submit', function(event) {
    event.preventDefault();
    const lpuID = document.getElementById('lpuID').value;
    const doctorID = document.getElementById('doctorID').value;
    const startDate = document.getElementById('startDate').value;
    const endDate = document.getElementById('endDate').value;
    const workDays = document.getElementById('workDays').value.split(',').map(day => parseInt(day));
    const weekParity = document.getElementById('weekParity').value;
    const dayParity = document.getElementById('dayParity').value;
    const startTime = document.getElementById('startTime').value;
    const endTime = document.getElementById('endTime').value;
    const appointmentDuration = document.getElementById('appointmentDuration').value;
    const officeNumber = document.getElementById('officeNumber').value;

    const scheduleData = {
        LpuID: parseInt(lpuID),
        DoctorID: parseInt(doctorID),
        ScheduleStartDate: startDate,
        ScheduleEndDate: endDate,
        ScheduleWorkDays: workDays,
        ScheduleWeekParity: weekParity,
        ScheduleDayParity: dayParity,
        ScheduleStartTime: startTime,
        ScheduleEndTime: endTime,
        ScheduleAppointmentDuration: parseInt(appointmentDuration),
        ScheduleOfficeNumber: parseInt(officeNumber),
    };

    fetch('http://localhost:8080/schedule', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(scheduleData)
    }).then(response => response.json())
    .then(data => alert('Расписание создано успешно'))
    .catch(error => alert('Ошибка: ' + error));
});

function getSchedule() {
    const doctorID = document.getElementById('getDoctorID').value;
    if (!doctorID) {
        alert('Пожалуйста, введите ID врача.');
        return;
    }
   
    const today = new Date();

    const dateRange = [];
    for (let i = 0; i < 14; i++) {
        const date = new Date(today);
        date.setDate(today.getDate() + i);
        dateRange.push(date.toISOString().split('T')[0]); 
    }

    const scheduleBody = document.getElementById('scheduleBody');
    scheduleBody.innerHTML = '';

    dateRange.forEach(date => {
        fetch(`http://localhost:8080/schedule/${doctorID}?date=${date}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Сетевая ошибка при получении данных');
                }
                return response.json();
            })
            .then(data => {
                displaySchedule(date, data);
            })
            .catch(error => {
                console.error(`Ошибка при получении расписания на ${date}:`, error);
                alert(`Ошибка при получении расписания на ${date}`);
            });
    });
}

function displaySchedule(date, data) {
    const scheduleBody = document.getElementById('scheduleBody');

    const row = document.createElement('tr');
    const dateHeader = document.createElement('th');
    dateHeader.textContent = date;
    row.appendChild(dateHeader);

    data.forEach(cell => {
        const cellInfo = document.createElement('td');
        cellInfo.textContent = `${cell.ScheduleCellTime}: ${cell.ScheduleStatus}`;
        row.appendChild(cellInfo);
    });

    scheduleBody.appendChild(row);
}

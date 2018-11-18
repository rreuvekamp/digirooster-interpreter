Date.prototype.getWeek = function() {
	let firstDay = new Date(this.getFullYear(), 0, 1)
	return Math.ceil((((this - firstDay) / 86400000) + firstDay.getDay() + 1) / 7)
}

function goToCurrentWeek() {
	let weekNumber = (new Date()).getWeek()
	location.href = "#week_" + weekNumber
}

document.querySelector("header a").addEventListener("click", function(e) {
	goToCurrentWeek()
	e.preventDefault()
})

goToCurrentWeek()

function highlightCurrent() {
	let d = (new Date())
	let m = d.getMinutes()
	if (m < 10) {
		m = "0" + m
	}
	now = d.getHours() + "" + m
	console.log(now)

	let week = document.querySelector("section[data-week='" + d.getWeek() + "']")

	offset = week.getAttribute("data-startat")
	end = week.getAttribute("data-endat")

	pixels = (now-offset)/2
	console.log(offset, now)

	let activities = week.querySelector("div[data-day='" + d.getDay() + "'] div.activities")

	if (pixels >= 0 && end-now > 0) {
		let line = document.createElement("div")
		line.classList.add("currenttimeline")
		line.style.marginTop = pixels + "px"
		activities.appendChild(line)
	}
}

highlightCurrent()

setTimeout(highlightCurrent, 30000) // 5 minutes

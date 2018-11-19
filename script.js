Date.prototype.getWeek = function() {
	let firstDay = new Date(this.getFullYear(), 0, 1)
	return Math.ceil((((this - firstDay) / 86400000) + firstDay.getDay() + 1) / 7)
}

function goToWeek(week) {
	console.log(week)
	if (!week) {
		week = (new Date()).getWeek()
	}
	location.href = "#week_" + week
}


goToWeek()

function highlightCurrent() {
	let d = (new Date())
	let m = d.getMinutes()
	if (m < 10) {
		m = "0" + m
	}
	now = d.getHours() + "" + m

	let week = document.querySelector("section[data-week='" + d.getWeek() + "']")

	offset = week.getAttribute("data-startat")
	end = week.getAttribute("data-endat")

	pixels = (now-offset)/2

	console.log(d.getDay(), parseInt(d.getDay()), d.getDay()-1)
	let activities = week.querySelector("div[data-day='" + (d.getDay()-1) + "'] div.activities")
	console.log(activities)

	if (activities && pixels >= 0 && end-now > 0) {
		let line = document.createElement("div")
		line.classList.add("currenttimeline")
		line.style.marginTop = pixels + "px"
		activities.appendChild(line)
	}
}

let titleEls = document.querySelectorAll("[title]")
for (let i = 0; i < titleEls.length; ++i) { 
	titleEls[i].addEventListener("click", function(e) {
		let title = titleEls[i].getAttribute("title")
		let content = titleEls[i].innerHTML
		titleEls[i].setAttribute("title", content)
		titleEls[i].innerHTML = title
	})
}

function determineDisplayedWeek() {
	let closest = 0;
	let distance = 0;

	let scroll = document.querySelector("html").scrollTop
	console.log(scroll)

	let els = document.querySelectorAll("section[data-week]")
	for (let i = 0; i < els.length; i++) {
		let diff = scroll-els[i].offsetTop
		if (diff < 0) {
			diff *= -1
		}

		if (distance == 0 || diff < distance) {
			distance = diff
			closest = parseInt(els[i].getAttribute("data-week"))
		}
	}
	console.log(closest)
	return closest
}
determineDisplayedWeek()

let displayedWeek = (new Date()).getWeek()
document.querySelector("button#left").addEventListener("click", function() {
	prevWeek()
})

function prevWeek() {
	let cur = determineDisplayedWeek()

	let els = document.querySelectorAll("section[data-week]")
	let last = -1;
	for (let i = 0; i < els.length; i++) {
		let thiss = els[i].getAttribute("data-week")
		console.log(thiss, cur, last)
		if (thiss == cur && last >= 0) {
			goToWeek(last)
			return
		}

		last = thiss
	}

	// Go to last if we're at the first
	goToWeek(els[els.length-1].getAttribute("data-week"))
}
document.querySelector("button#right").addEventListener("click", function() {
	nextWeek()
})

function nextWeek(){
	let cur = determineDisplayedWeek()

	let els = document.querySelectorAll("section[data-week]")
	let last = 0;
	for (let i = 0; i < els.length; i++) {
		let thiss = els[i].getAttribute("data-week")
		if (last == cur) {
			goToWeek(thiss)
			return
		}

		last = thiss
	}

	// Go to first if we're al the last
	goToWeek(els[0].getAttribute("data-week"))
}
document.querySelector("button#currentweek").addEventListener("click", function(e) {
	goToWeek()
	e.preventDefault()
})
document.addEventListener("keydown", function(e) {
	console.log(e.keyCode)
	switch (e.keyCode) {
	case 37:
		prevWeek()
		break
	case 39:
		nextWeek()
		break
	}
})


highlightCurrent()

setTimeout(highlightCurrent, 30000) // 5 minutes

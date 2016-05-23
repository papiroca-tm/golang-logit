package logit

import (
	. "github.com/smartystreets/goconvey/convey"
	"regexp"
	"testing"
	"time"
)

func TestTRACE(t *testing.T) {
	Convey("Функция не должна упасть по panic", t, func() {
		So(func() { TRACE("someLogText", "someLogContext") }, ShouldNotPanic)
	})
}

func TestINFO(t *testing.T) {
	Convey("Функция не должна упасть по panic", t, func() {
		So(func() { INFO("someLogText", "someLogContext") }, ShouldNotPanic)
	})
}

func TestWARN(t *testing.T) {
	Convey("Функция не должна упасть по panic", t, func() {
		So(func() { WARN("someLogText", "someLogContext") }, ShouldNotPanic)
	})
}

func TestERROR(t *testing.T) {
	Convey("Функция не должна упасть по panic", t, func() {
		So(func() { ERROR("someLogText", "someLogContext", "someErrorCode") }, ShouldNotPanic)
	})
}

func TestTimeToStr(t *testing.T) {
	Convey("Конвертация времени в строку", t, func() {
		tStr := timeToStr(time.Now())
		Convey(tStr+" t time должно быть преобразованно в формат 02.01.2006 15:04:05", func() {
			ok, _ := regexp.MatchString("[0-3][0-9]\\.[0-1][0-9]\\.[0-9][0-9][0-9][0-9] [0-9][0-9]:[0-9][0-9]:[0-9][0-9]", tStr)
			So(ok, ShouldBeTrue)
		})
	})
}

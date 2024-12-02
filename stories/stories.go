package stories

import (
	"net/url"
	"reflect"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofiber/fiber/v2/log"
)

type CurrentStory struct {
	ID        string    `faker:"int,unique"`
	Title     string    `faker:"sentence"`
	Published time.Time `faker:"ttime"`
	Posted    time.Time `faker:"ttime"`
	URL       url.URL   `faker:"uurl"`
	Comments  url.URL   `faker:"uurl"`
	Source    string    `faker:"oneof:hn,ls"`
}

func init() {
	faker.AddProvider("uurl", func(v reflect.Value) (any, error) {
		u, e := url.Parse(faker.URL())
		if e != nil {
			return url.URL{}, e
		} else {
			return *u, nil
		}
	})
	faker.AddProvider("ttime", func(v reflect.Value) (interface{}, error) {
		ddate := faker.Timestamp()
		return time.Parse(faker.BaseDateFormat+" "+faker.TimeFormat, ddate)
	})
}

func TopStories() []CurrentStory {
	stories := make([]CurrentStory, 10)
	for i := range stories {

		story := CurrentStory{}

		err := faker.FakeData(&story)
		if err != nil {
			log.Error(err)
		}
		stories[i] = story
	}
	return stories
}

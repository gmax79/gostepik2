package provider

import (
	"encoding/json"
)

type settings struct {
	Database dbsettings `json:"database"`
}

type dbsettings struct {
	Host     string `json:"host"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type content struct {
	db         dbsettings
	allCources map[string]courseData
}

type courseData struct {
	description string
	name        string
	lessons     map[string]lessonData
}

type lessonData struct {
	description string
	name        string
}

// Intialize - run provider, read data, setup db connections
func create(settingsJSON []byte) (CourseHandler, error) {
	var database settings
	if err := json.Unmarshal(settingsJSON, &database); err != nil {
		return nil, err
	}
	instance := &content{db: database.Database}
	if err := instance.checkAndInitDatabase(); err != nil {
		return nil, err
	}

	instance.allCources = map[string]courseData{
		"gobeginner": courseData{
			name: "Go", description: "Базовые знания Go",
			lessons: map[string]lessonData{
				"btd":     lessonData{name: "Базовые типы данных", description: "БТД - byte, string, int, uint"},
				"structs": lessonData{name: "Структуры", description: "Структуры - объекты в Go"},
			},
		},
		"cpp11": courseData{
			name: "С++11", description: "Изучение стандарта C++11",
			lessons: map[string]lessonData{
				"cppchanges":   lessonData{name: "Обзор изменений в С++11", description: "ОИ"},
				"movesemantic": lessonData{name: "Семантика перемещения", description: "Move"},
			},
		},
	}
	return instance, nil
}

func (u *content) GetCources() []Course {
	list := make([]Course, 0, len(u.allCources))
	for id, c := range u.allCources {
		list = append(list, Course{ID: id, Name: c.name, Description: c.description, LessonsCount: len(c.lessons)})
	}
	return list
}

func (u *content) GetLessons(courceID string) ([]Lesson, bool) {
	if c, ok := u.allCources[courceID]; ok {
		list := make([]Lesson, 0, len(c.lessons))
		for id, lesson := range c.lessons {
			list = append(list, Lesson{LessonIndex: LessonIndex{CourseID: courceID, LessonID: id}, Name: lesson.name, Description: lesson.description})
		}
		return list, true
	}
	return []Lesson{}, false
}

func (u *content) GetLesson(id LessonIndex) (Lesson, bool) {
	if c, ok := u.allCources[id.CourseID]; ok {
		if l, ok2 := c.lessons[id.LessonID]; ok2 {
			return Lesson{Name: l.name, Description: l.description, LessonIndex: id}, true
		}
	}
	return Lesson{}, false
}

func (u *content) SaveReport(state LessonReport) {

}

package provider

type LessonIndex struct {
	CourseID string
	LessonID string
}

type Lesson struct {
	LessonIndex
	Name        string
	Description string
}

type Course struct {
	ID           string
	Name         string
	Description  string
	LessonsCount int
}

type LessonReport struct {
	LessonIndex
	Login   string
	TmpFile string
	State   int
}

/*type User struct {
	Login    string
	Password string
	State    []LessonState
}*/

// CourseHandler - main api interface
type CourseHandler interface {
	GetCources() []Course
	GetLessons(courceID string) ([]Lesson, bool)
	GetLesson(id LessonIndex) (Lesson, bool)
	SaveReport(rep LessonReport)
}

// Create - create instance of code, controlled courses information
func Create(jsonSettings []byte) (CourseHandler, error) {
	return create(jsonSettings)
}
